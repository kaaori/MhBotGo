package profiler

import (
	"log"
	"strconv"
	"time"
)

var (
	debug     = false
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
	log.Println(msg + " -- Time taken: " + strconv.Itoa(int(time.Since(startTime).Seconds())) + " seconds~")
	startTime = time.Now()
}

func SecondsSinceAsString() string {
	if !debug {
		return ""
	}
	startTime = time.Now()
	return strconv.Itoa(int(time.Since(startTime).Seconds()))
}
