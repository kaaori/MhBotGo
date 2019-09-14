package chrome

import (
	"context"
	"io/ioutil"
	"log"
	"os"
	"sync"

	"github.com/chromedp/cdproto/emulation"
	"github.com/chromedp/cdproto/page"
	"github.com/chromedp/chromedp"
)

var (
	ctx chromedp.Context
)

// Init : Initialises a chrome instance
func Init() (context.Context, context.CancelFunc) {
	// create context
	ctx, cancel := chromedp.NewContext(context.Background())
	return ctx, cancel
}

// TakeScreenshot : Takes a screenshot of the given url
func TakeScreenshot(w int64, h int64, channelID string) {
	ctx, cancel := Init()
	defer cancel()

	path, _ := os.Getwd()

	// capture entire browser viewport, returning png with quality=90
	var buf []byte
	if err := chromedp.Run(ctx, fullScreenshot("file:///"+path+"/web/schedule-parsed.html",
		120, &buf, w, h)); err != nil {
		log.Fatal(err)
	}

	var todayBuf []byte
	if err := chromedp.Run(ctx, fullScreenshot("file:///"+path+"/web/today-parsed.html",
		120, &buf, w, h)); err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile("schedule-chan-"+channelID+".png", buf, 0644); err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile("today-chan-"+channelID+".png", todayBuf, 0644); err != nil {
		log.Fatal(err)
	}
}

// TakeScreenshotTargeted : Takes a screenshot of the given url by target
func TakeScreenshotTargeted(w int64, h int64, element string, fileName string, urlString string, waitgroup ...*sync.WaitGroup) {
	if len(waitgroup) > 0 {
		defer waitgroup[0].Done()
	}

	ctx, cancel := Init()
	defer cancel()

	var buf []byte
	emulation.SetScrollbarsHidden(true)
	emulation.SetDeviceMetricsOverride(w, h, 1, false)
	if err := chromedp.Run(ctx, elementScreenshot(urlString,
		element, &buf, w, h)); err != nil {
		log.Fatal(err)
	}

	if err := ioutil.WriteFile(fileName, buf, 0644); err != nil {
		log.Fatal(err)
	}
}

// elementScreenshot takes a screenshot of a specific element.
func elementScreenshot(urlstr, sel string, res *[]byte, w int64, h int64) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.WaitVisible(sel, chromedp.ByID),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible, chromedp.ByID),
	}
}

// fullScreenshot takes a screenshot of the entire browser viewport.
//
// Liberally copied from puppeteer's source.
//
// Note: this will override the viewport emulation settings.
func fullScreenshot(urlstr string, quality int64, res *[]byte, w int64, h int64) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			// width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))
			width, height := w, h
			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, .75, false).
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
					Width:  float64(width),
					Height: float64(height),
					Scale:  1,
				}).Do(ctx)
			if err != nil {
				return err
			}
			return nil
		}),
	}
}
