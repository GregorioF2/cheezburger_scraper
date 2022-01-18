package images

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"math"
	"os"
	"sort"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/dom"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	config "propper/configs"
	logger "propper/lib/logger"
	sem "propper/lib/semaphore"

	. "propper/types/errors"
)

func urlOfPage(url string, page int) string {
	if page <= 1 {
		return url
	}
	return fmt.Sprintf("%s/page/%d", url, page)
}

func extractSrcFromNode(node *cdp.Node) string {
	src, exists := node.Attribute("data-src")
	if exists {
		return src

	}
	src, exists = node.Attribute("src")
	if exists {
		return src
	}
	return ""
}

func downloadImages(ctx context.Context, urls []string, path string) error {
	var requestInProgressWG sync.WaitGroup
	var currReqId network.RequestID

	chromedp.ListenTarget(ctx, func(v interface{}) {
		switch ev := v.(type) {
		case *network.EventRequestWillBeSent:
			currReqId = ev.RequestID
		case *network.EventLoadingFinished:
			if ev.RequestID == currReqId {
				requestInProgressWG.Done()
			}
		}
	})
	var waitForActions sync.WaitGroup
	waitForActions.Add(1)
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			defer waitForActions.Done()
			for i, url := range urls {
				requestInProgressWG.Add(1)
				err := chromedp.Navigate(url).Do(ctx)
				if err != nil {
					return &ConnectionError{Err: fmt.Sprintf("Error downloading image from url: %s", url), RawError: err}
				}
				requestInProgressWG.Wait()
				buf, err := network.GetResponseBody(currReqId).Do(ctx)
				if err != nil {
					return &InternalServerError{Err: "Unexpected error downloading image.", RawError: err}
				}
				if err := ioutil.WriteFile(fmt.Sprintf("%s/%d.jpg", path, i+1), buf, 0644); err != nil {
					return &InternalServerError{Err: "Unexpected error writing image locally.", RawError: err}
				}
			}
			return nil
		}),
	)

	waitForActions.Wait()
	logger.Log("Finished downloading the images")
	return err
}

func getImagesURLS(ctx context.Context, amount, threads int) ([]string, error) {
	logger.Log("Start getting the urls")

	if amount < 1 {
		return nil, &InvalidParametersError{Err: "amount must be greater or equal than 1."}
	}
	if threads < 1 || threads > 5 {
		return nil, &InvalidParametersError{Err: "threads must be greater or equal than 1, and lesser or equal than 5."}
	}

	var maxConcurrentThreads int = threads
	maxTotalQueries := int(math.Ceil(float64(amount) / float64(config.MIN_CARDS_PER_PAGE)))
	if maxConcurrentThreads > maxTotalQueries {
		maxConcurrentThreads = maxTotalQueries
	}
	logger.Log(fmt.Sprintf("max concurrent threads: %d", maxConcurrentThreads))
	logger.Log(fmt.Sprintf("max total queries: %d", maxTotalQueries))
	semConcurrentThreads := sem.NewCustomSemaphore(maxConcurrentThreads)
	defer semConcurrentThreads.Close()

	resMap := sync.Map{}
	var imageUrls []string

	errs := make(chan error, maxConcurrentThreads)

	var wg sync.WaitGroup
	var tabs []context.Context
	for i := 0; i < maxConcurrentThreads; i += 1 {
		newCtx, _ := chromedp.NewContext(ctx)
		tabs = append(tabs, newCtx)
	}
	resolvedUrls := 0
	getNodesOfPage := func(cc context.Context, page int) {
		if err := chromedp.Run(cc,
			chromedp.ActionFunc(func(cc context.Context) error {
				defer wg.Done()
				defer semConcurrentThreads.Signal()

				logger.Log(fmt.Sprintf("Go routine for page %d started", page))
				var localNodes []*cdp.Node
				url := urlOfPage(config.SITE_URL, page)
				localUrls := []string{}

				err := chromedp.Navigate(url).Do(cc)
				if err != nil {
					return &ConnectionError{Err: fmt.Sprintf("Error connecting to URL (%s)", url), RawError: err}
				}
				// wait to load resources
				err = chromedp.Sleep(time.Second * time.Duration(config.SLEEP_TIME)).Do(cc)
				if err != nil {
					return &InternalServerError{Err: err.Error(), RawError: err}
				}
				_, resultCount, err := dom.PerformSearch(config.CARD_IMG_SELECTOR).Do(cc)
				if err != nil {
					return &InternalServerError{Err: err.Error(), RawError: err}
				}
				if resultCount == 0 {
					return &NotFoundError{Err: fmt.Sprintf("Couldn't find any images on url(%s)", url)}
				}
				err = chromedp.Nodes(config.CARD_IMG_SELECTOR, &localNodes, chromedp.BySearch).Do(cc)
				if err != nil {
					return &InternalServerError{Err: "Unexpected error selecting nodes", RawError: err}
				}

				for _, node := range localNodes {
					localUrls = append(localUrls, extractSrcFromNode(node))
				}
				resMap.Store(page, localUrls)
				resolvedUrls += len(localNodes)
				logger.Log(fmt.Sprintf("Go routine for page %d finished", page))
				return nil
			}),
		); err != nil {
			errs <- err
			return
		}
	}

	logger.Log("Start routines")
	for i := 0; i < maxTotalQueries; i += 1 {
		if resolvedUrls+semConcurrentThreads.CurrentlyRunning()*config.MIN_CARDS_PER_PAGE > amount {
			logger.Log("Preemptive break on starting new routines")
			break
		}
		wg.Add(1)
		semConcurrentThreads.Take()
		go getNodesOfPage(tabs[i%maxConcurrentThreads], i+1)
	}
	wg.Wait()
	close(errs)
	if len(errs) > 0 {
		return nil, <-errs
	}

	keys := []int{}
	resMap.Range(func(key interface{}, value interface{}) bool {
		keys = append(keys, key.(int))
		return true
	})
	sort.Ints(keys)
	for _, page := range keys {
		urls, ok := resMap.Load(page)
		if !ok {
			return nil, &InternalServerError{Err: "No results retrieved for one of the pages"}
		}
		imageUrls = append(imageUrls, urls.([]string)...)
	}
	if amount > len(imageUrls) {
		return nil, &BadRequestError{Err: "Not enough images to meet the amount"}
	}
	logger.Log("Finished getting the urls")
	return imageUrls[0:amount], nil
}

// Given a number of images and number of threads to use. It takes care of coordinating
// the search and download of the images of the specified site in configs.
// It returns the urls of the downloaded images.
func GetImages(amount, threads int) ([]string, error) {
	// create context
	maintCtx, _ := chromedp.NewContext(
		context.Background(),
		chromedp.WithLogf(log.Printf),
	)
	imagesCtx, _ := chromedp.NewContext(
		maintCtx,
		chromedp.WithLogf(log.Printf),
	)
	// create a timeout as a safety net to prevent any infinite wait loops
	maintCtx, cancel := context.WithTimeout(maintCtx, time.Duration(config.TIMEOUT)*time.Second)
	defer cancel()

	imageUrls, err := getImagesURLS(maintCtx, amount, threads)
	if err != nil {
		return nil, err
	}

	folderName := time.Now().UTC().Format("2006_01_02 15:04:05")
	saveDirectoryPath := fmt.Sprintf("%s/%s", config.DOWNLOADS_SAVE_DIR, folderName)
	err = os.Mkdir(saveDirectoryPath, 0755)
	if err != nil {
		return nil, &InternalServerError{Err: err.Error(), RawError: err}
	}

	err = downloadImages(imagesCtx, imageUrls, saveDirectoryPath)
	if err != nil {
		return nil, err
	}
	return imageUrls, nil
}
