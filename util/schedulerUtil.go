package util

import (
	"math"
	"strconv"
	"time"
)

// GetRoundedMinutesTilEvent : Gets minutes rounded up from time as string
func GetRoundedMinutesTilEvent(startTime time.Time) string {
	return strconv.Itoa(int(math.Ceil(time.Until(startTime).Minutes())))
}
