package util

import (
	"time"
)

var (
	// ServerTz : Server tz and offset
	ServerTz string

	// ServerLocOffset : The int64 offset from unix time of the server
	ServerLocOffset int64

	// ServerLoc : Locale of the current server
	ServerLoc *time.Location
)

func init() {
	ServerLoc = time.Local
	serverTz, locOffsetInt := time.Now().In(ServerLoc).Zone()
	ServerLocOffset = int64(locOffsetInt)
	ServerTz = serverTz
}

// GetCurrentWeekFromMondayAsTime : Gets the time object representing the current week starting @ monday
func GetCurrentWeekFromMondayAsTime() time.Time {
	_, week := time.Now().ISOWeek()
	return FirstDayOfISOWeek(time.Now().Year(), week, ServerLoc)
}

// func getEstOffset() int64 {
// 	_, offset := time.Now().In(ServerLoc).Zone()
// 	return int64(offset)
// }

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

	return date.In(ServerLoc)
}
