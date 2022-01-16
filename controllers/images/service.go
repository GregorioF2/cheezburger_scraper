package images

import (
	"context"
	"errors"
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

func advancePage(url string, page int) string {
	if page <= 1 {
		return url
	}
	return fmt.Sprintf("%s/page/%d", url, page)
}

func GetImagesURLS(ctx context.Context, ammount int) ([]string, error) {
	var nodes []*cdp.Node
	var imageUrls []string
	var pageNumber int = 1

	var waitForActions sync.WaitGroup
	waitForActions.Add(1)
	logger.Log(fmt.Sprintf("Ammount to meet: %d", ammount))
	logger.Log("Start getting the urls")
	if err := chromedp.Run(ctx,
		chromedp.Navigate(config.SITE_URL),
		chromedp.ActionFunc(func(ctx context.Context) error {
			logger.Log("Start action to select nodes")
			for len(nodes) <= ammount {
				var localNodes []*cdp.Node
				err := chromedp.Nodes(config.CARD_IMG_SELECTOR, &localNodes, chromedp.BySearch).Do(ctx)
				if err != nil {
					return err
				}
				logger.Log(fmt.Sprintf("Nodes on page %d: %d", pageNumber, len(localNodes)))
				nodesLeft := ammount - len(nodes)

				if len(localNodes) >= nodesLeft {
					nodes = append(nodes, localNodes[0:nodesLeft]...)
					return nil
				}
				nodes = append(nodes, localNodes...)
				pageNumber += 1
				err = chromedp.Navigate(advancePage(config.SITE_URL, pageNumber)).Do(ctx)
				if err != nil {
					return err
				}
			}
			return nil
		}),
		chromedp.ActionFunc(func(ctx context.Context) error {
			logger.Log("Start action to filter nodes attributes")
			for _, node := range nodes {
				src, exists := node.Attribute("data-src")
				if exists {
					imageUrls = append(imageUrls, src)
					continue
				}
				src, exists = node.Attribute("src")
				if exists {
					imageUrls = append(imageUrls, src)
					continue
				}
				return errors.New("Image does not have src url")
			}
			waitForActions.Done()
			return nil
		}),
	); err != nil {
		return nil, err
	}

	waitForActions.Wait()
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
