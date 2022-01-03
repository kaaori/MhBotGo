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

// GetCurrentDayForSchedule : Gets the class for setting the active day on the schedule
func GetCurrentDayForSchedule(day time.Weekday) string {
	if time.Now().Weekday() == day {
		return "active"
	}
	return ""
}

func init() {
	ServerLoc = time.Local
	serverTz, locOffsetInt := time.Now().In(ServerLoc).Zone()
	ServerLocOffset = int64(locOffsetInt)
	ServerTz = serverTz
}

// GetCurrentWeekFromMondayAsTime : Gets the time object representing the current week starting @ monday
func GetCurrentWeekFromMondayAsTime() time.Time {
	year, week := time.Now().ISOWeek()

	date := time.Date(year, 0, 0, 0, 0, 0, 0, ServerLoc)
	isoYear, isoWeek := date.ISOWeek()
	for date.Weekday() != time.Monday { // iterate back to Monday
		date = date.AddDate(0, 0, -1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoYear < year { // iterate forward to the first day of the first week
		date = date.AddDate(0, 0, 1)
		isoYear, isoWeek = date.ISOWeek()
	}
	for isoWeek < week { // iterate forward to the first day of the given week
		date = date.AddDate(0, 0, 1)
		_, isoWeek = date.ISOWeek()
	}
	return date
}
