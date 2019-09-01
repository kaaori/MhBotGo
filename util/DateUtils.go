package util

import (
	"time"
)

var (
	EstLoc, _    = time.LoadLocation("America/New_York")
	EstLocOffset = int64(getEstOffset())
)

func getEstOffset() int {
	_, offset := time.Now().In(EstLoc).Zone()
	return offset
}

// FirstDayOfISOWeek : Gets the time object for the first date in a given week (extracted from time.Now())
func FirstDayOfISOWeek(year int, week int, timezone *time.Location) time.Time {
	date := time.Date(year, 0, 0, 0, 0, 0, 0, timezone)
	isoYear, isoWeek := date.ISOWeek()

	// iterate back to Monday
	for date.Weekday() != time.Monday {
		date = date.AddDate(0, 0, -1)
		isoYear, isoWeek = date.ISOWeek()
	}

	// iterate forward to the first day of the first week
	for isoYear < year {
		date = date.AddDate(0, 0, 7)
		isoYear, isoWeek = date.ISOWeek()
	}

	// iterate forward to the first day of the given week
	for isoWeek < week {
		date = date.AddDate(0, 0, 7)
		isoYear, isoWeek = date.ISOWeek()
	}

	return date.In(EstLoc)
}
