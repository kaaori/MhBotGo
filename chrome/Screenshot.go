package chrome

import (
	"context"
	"io/ioutil"
	"log"
	"os"

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
func TakeScreenshot() {
	ctx, cancel := Init()
	defer cancel()

	path, _ := os.Getwd()

	// capture entire browser viewport, returning png with quality=90
	var buf []byte
	if err := chromedp.Run(ctx, elementScreenshot("file:///"+path+"/web/schedule-parsed.html",
		"#main", &buf)); err != nil {
		log.Fatal(err)
	}
	if err := ioutil.WriteFile("schedule.png", buf, 0644); err != nil {
		log.Fatal(err)
	}
}

// elementScreenshot takes a screenshot of a specific element.
func elementScreenshot(urlstr, sel string, res *[]byte) chromedp.Tasks {
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
func fullScreenshot(urlstr string, quality int64, res *[]byte) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		chromedp.ActionFunc(func(ctx context.Context) error {
			// get layout metrics
			_, _, contentSize, err := page.GetLayoutMetrics().Do(ctx)
			if err != nil {
				return err
			}

			// width, height := int64(math.Ceil(contentSize.Width)), int64(math.Ceil(contentSize.Height))
			width, height := int64(1280), int64(720)

			// force viewport emulation
			err = emulation.SetDeviceMetricsOverride(width, height, 1, false).
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
