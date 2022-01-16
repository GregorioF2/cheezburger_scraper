package images

import (
	"context"
	"fmt"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"time"

	"github.com/chromedp/cdproto/cdp"
	"github.com/chromedp/cdproto/network"
	"github.com/chromedp/chromedp"

	config "propper/configs"
	types "propper/types"
	logger "propper/types/logger"
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
			waitForActions.Done()
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

func GetImagesURLS(ctx context.Context, ammount int) ([]string, error) {
	var maxConcurrentThreads int = 5
	if maxConcurrentThreads > ammount/10 {
		maxConcurrentThreads = ammount / 10
	}
	maxTotalThreads := ammount / 10

	var nodes []*cdp.Node
	var imageUrls []string

	logger.Log(fmt.Sprintf("Ammount to meet: %d", ammount))
	logger.Log("Start getting the urls")

	out := make(chan types.PageNode, maxConcurrentThreads*30)
	errs := make(chan error, maxConcurrentThreads)
	semConcurrentThreads := make(chan int, maxConcurrentThreads)
	var wg sync.WaitGroup
	var tabs []context.Context
	for i := 0; i < maxConcurrentThreads; i += 1 {
		newCtx, _ := chromedp.NewContext(ctx)
		tabs = append(tabs, newCtx)
	}

	getNodesOfPage := func(cc context.Context, page int) {
		if err := chromedp.Run(cc,
			chromedp.ActionFunc(func(cc context.Context) error {
				logger.Log(fmt.Sprintf("Go routine for page %d started", page))
				var localNodes []*cdp.Node
				url := urlOfPage(config.SITE_URL, page)
				defer wg.Done()

				err := chromedp.Navigate(url).Do(cc)
				if err != nil {
					return err
				}
				logger.Log(fmt.Sprintf("Bk1 goroutine: %d", page))
				err = chromedp.Nodes(config.CARD_IMG_SELECTOR, &localNodes, chromedp.BySearch).Do(cc)
				if err != nil {
					return err
				}
				logger.Log(fmt.Sprintf("Nodes on page %d: %d", page, len(localNodes)))

				for _, node := range localNodes {
					out <- types.PageNode{
						Node: node,
						Page: page,
						Url:  extractSrcFromNode(node),
					}
				}
				logger.Log(fmt.Sprintf("Go routine for page %d finished", page))
				<-semConcurrentThreads
				return nil
			}),
		); err != nil {
			errs <- err
			<-semConcurrentThreads
			return
		}
	}

	logger.Log("Start sendinding go ruoutines")
	for i := 0; i < maxTotalThreads; i += 1 {
		wg.Add(1)
		go getNodesOfPage(tabs[i], i+1)
		logger.Log("Before sem")
		semConcurrentThreads <- i + 1
		logger.Log("after sem")
	}
	logger.Log("Before wait")
	wg.Wait()
	close(out)
	close(errs)
	logger.Log("After wait")

	for val := range out {
		nodes = append(nodes, val.Node)
		imageUrls = append(imageUrls, val.Url)
	}
	logger.Log(fmt.Sprintf("Total ammount of nodes: %d ", len(nodes)))
	logger.Log("Finished getting the urls")
	return imageUrls, nil
}

func GetImages(ammount int) error {
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

	imageUrls, err := GetImagesURLS(maintCtx, ammount)
	if err != nil {
		return err
	}

	folderName := time.Now().UTC().Format("2006_01_02 15:04:05")
	saveDirectoryPath := fmt.Sprintf("downloads/%s", folderName)
	err = os.Mkdir(saveDirectoryPath, 0755)
	if err != nil {
		return err
	}

	err = DownloadImages(imagesCtx, imageUrls, saveDirectoryPath)
	if err != nil {
		return err
	}
	return nil
}
