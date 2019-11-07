package profiler

import (
	"strconv"
	"time"

	logging "github.com/kaaori/mhbotgo/log"
)

var (
	debug     = false
	log       = logging.NewLog()
	startTime time.Time
)

func Start() {
	if !debug {
		return
	}
	startTime = time.Now()
}

func StopAndPrintSeconds(msg string) {
	if !debug {
		return
	}
	log.Trace(msg + " -- Time taken: " + strconv.Itoa(int(time.Since(startTime).Seconds())) + " seconds~")
	startTime = time.Now()
}

func SecondsSinceAsString() string {
	if !debug {
		return ""
	}
	startTime = time.Now()
	return strconv.Itoa(int(time.Since(startTime).Seconds()))
}
