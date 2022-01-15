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
	return err
}

func GetImagesURLS(ctx context.Context) ([]string, error) {
	var nodes []*cdp.Node
	var imageUrls []string

	var waitForActions sync.WaitGroup
	waitForActions.Add(1)
	if err := chromedp.Run(ctx,
		chromedp.Navigate(config.SITE_URL),
		chromedp.Nodes(".mu-post.mu-thumbnail > img", &nodes, chromedp.BySearch),
		chromedp.ActionFunc(func(ctx context.Context) error {
			for _, node := range nodes[0:10] {
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
	return imageUrls, nil
}

func GetImages() error {
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
	maintCtx, cancel := context.WithTimeout(maintCtx, 60*time.Second)
	defer cancel()

	imageUrls, err := GetImagesURLS(maintCtx)
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
