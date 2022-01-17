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
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	config "propper/configs"
	logger "propper/lib/logger"
	sem "propper/lib/semaphore"
)

func DownloadImages(ctx context.Context, urls []string, path string) error {
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
	logger.Log("Start download the images")
	err := chromedp.Run(ctx,
		chromedp.ActionFunc(func(ctx context.Context) error {
			defer waitForActions.Done()
			for i, url := range urls {
				requestInProgressWG.Add(1)
				err := chromedp.Navigate(url).Do(ctx)
				if err != nil {
					return err
				}
				requestInProgressWG.Wait()
				buf, err := network.GetResponseBody(currReqId).Do(ctx)
				if err != nil {
					return err
				}
				if err := ioutil.WriteFile(fmt.Sprintf("%s/%d.jpg", path, i+1), buf, 0644); err != nil {
					return err
				}
			}
			return nil
		}),
	)

	waitForActions.Wait()
	logger.Log("Finished downloading the images")
	return err
}

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

func GetImagesURLS(ctx context.Context, ammount, threads int) ([]string, error) {
	logger.Log("Start getting the urls")
	var maxConcurrentThreads int = threads
	maxTotalThreads := int(math.Ceil(float64(ammount) / float64(config.MIN_CARDS_PER_PAGE)))
	if maxConcurrentThreads > maxTotalThreads {
		maxConcurrentThreads = maxTotalThreads
	}
	logger.Log(fmt.Sprintf("max concurrent threads: %d", maxConcurrentThreads))
	logger.Log(fmt.Sprintf("maxTotalThreads: %d", maxTotalThreads))
	semConcurrentThreads := sem.NewCustomSemaphore(maxConcurrentThreads)
	defer semConcurrentThreads.Close()

	resMap := map[int][]string{}
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
				logger.Log(fmt.Sprintf("Go routine for page %d started", page))
				var localNodes []*cdp.Node
				url := urlOfPage(config.SITE_URL, page)
				defer wg.Done()
				defer semConcurrentThreads.Signal()
				resMap[page] = []string{}

				err := chromedp.Navigate(url).Do(cc)
				if err != nil {
					return err
				}

				err = chromedp.Nodes(config.CARD_IMG_SELECTOR, &localNodes, chromedp.BySearch).Do(cc)
				if err != nil {
					return err
				}

				for _, node := range localNodes {
					resMap[page] = append(resMap[page], extractSrcFromNode(node))
				}
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
	for i := 0; i < maxTotalThreads; i += 1 {
		wg.Add(1)
		semConcurrentThreads.Take()
		go getNodesOfPage(tabs[i%maxConcurrentThreads], i+1)
	}
	wg.Wait()
	close(errs)
	if len(errs) > 0 {
		return nil, <-errs
	}

	keys := make([]int, len(resMap))
	i := 0
	for k := range resMap {
		keys[i] = k
		i++
	}
	sort.Ints(keys)
	for _, page := range keys {
		imageUrls = append(imageUrls, resMap[page]...)
	}
	logger.Log("Finished getting the urls")
	return imageUrls[0:ammount], nil
}

func GetImages(ammount, threads int) ([]string, error) {
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

	imageUrls, err := GetImagesURLS(maintCtx, ammount, threads)
	if err != nil {
		return nil, err
	}

	folderName := time.Now().UTC().Format("2006_01_02 15:04:05")
	saveDirectoryPath := fmt.Sprintf("%s/%s", config.DOWNLOADS_SAVE_DIR, folderName)
	err = os.Mkdir(saveDirectoryPath, 0755)
	if err != nil {
		return nil, err
	}

	err = DownloadImages(imagesCtx, imageUrls, saveDirectoryPath)
	if err != nil {
		return nil, err
	}
	return imageUrls, nil
}
