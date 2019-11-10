package chrome

import (
	"context"
	"io/ioutil"
	"math"
	"sync"

	logging "github.com/kaaori/mhbotgo/log"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

var (
	ctx chromedp.Context
	mtx sync.Mutex
	log = logging.NewLog()
)

// Init : Initialises a chrome instance
func Init() (context.Context, context.CancelFunc) {
	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	return ctx, cancel
}

// TakeScreenshot : Takes a screenshot of the given url by target or full page
func TakeScreenshot(w int64, h int64, element string, fileName string, urlString string, isTargeted bool, waitgroup ...*sync.WaitGroup) {
	if len(waitgroup) > 0 {
		defer waitgroup[0].Done()
	}

	ctx, cancel := chromedp.NewContext(context.Background())
	// defer ctx.Done()
	defer cancel()

	var buf []byte
	if isTargeted {
		if err := chromedp.Run(ctx, elementScreenshot(urlString,
			element, &buf)); err != nil {
			log.Error("Error setting device metrics: ", err)
		}
	} else {
		if err := chromedp.Run(ctx, fullScreenshot(urlString,
			100, &buf)); err != nil {
			log.Error("Error setting device metrics: ", err)
		}
	}

	if err := ioutil.WriteFile(fileName, buf, 0644); err != nil {
		log.Error("Error writing file: ", err)
	}
}

// elementScreenshot takes a screenshot of a specific element.
func elementScreenshot(urlstr, sel string, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		// chromedp.ScrollIntoView(sel),
		// chromedp.Sleep(1 * time.Second),
		chromedp.WaitVisible(sel, chromedp.ByID),
		// chromedp.
		chromedp.Screenshot(sel, res, chromedp.NodeVisible, chromedp.ByID),
	}
}

// fullScreenshot takes a screenshot of the entire browser viewport.
//
// Liberally copied from puppeteer's source.
//
// Note: this will override the viewport emulation settings.
func fullScreenshot(urlstr string, quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, 1, true).
				WithScreenOrientation(&emulation.ScreenOrientation{
					Type:  emulation.OrientationTypePortraitPrimary,
					Angle: 0,
				}).
				Do(ctx)
			if err != nil {
				return err
			}

			// capture screenshot
			*res, err = page.CaptureScreenshot().
				WithQuality(quality).
				WithClip(&page.Viewport{
					X:      contentSize.X,
					Y:      contentSize.Y,
					Width:  contentSize.Width,
					Height: contentSize.Height,
					Scale:  1,
				}).Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
	}
}
