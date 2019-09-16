package chrome

import (
	"context"
	"io/ioutil"
	"sync"
	"time"

	logging "github.com/kaaori/mhbotgo/log"

	"github.com/chromedp/cdproto/emulation"
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

// TakeScreenshotTargeted : Takes a screenshot of the given url by target
func TakeScreenshotTargeted(w int64, h int64, element string, fileName string, urlString string, waitgroup ...*sync.WaitGroup) {
	if len(waitgroup) > 0 {
		defer waitgroup[0].Done()
	}

	ctx, cancel := Init()
	defer ctx.Done()
	defer cancel()

	var buf []byte
	emulation.SetScrollbarsHidden(true)
	emulation.SetDeviceMetricsOverride(w, h, 1, false)
	if err := chromedp.Run(ctx, elementScreenshot(urlString,
		element, &buf, w, h)); err != nil {
		log.Error("Error setting device metrics: ", err)
	}

	if err := ioutil.WriteFile(fileName, buf, 0644); err != nil {
		log.Error("Error writing file: ", err)
	}
}

// elementScreenshot takes a screenshot of a specific element.
func elementScreenshot(urlstr, sel string, res *[]byte, w int64, h int64) chromedp.Tasks {
	return chromedp.Tasks{
		chromedp.Navigate(urlstr),
		// chromedp.ScrollIntoView(sel),
		chromedp.WaitVisible(sel, chromedp.ByID),
		chromedp.Sleep(500 * time.Millisecond),
		chromedp.Screenshot(sel, res, chromedp.NodeVisible, chromedp.ByID),
	}
}
