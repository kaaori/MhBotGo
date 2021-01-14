package util

import (
	"time"

	"github.com/snabb/isoweek"
	config "github.com/spf13/viper"
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
	_, week := time.Now().ISOWeek()
	return isoweek.StartTime(time.Now().Year()+config.GetInt("weekStartOffset"), week, ServerLoc)
	// return FirstDayOfISOWeek(time.Now().Year(), week, ServerLoc)
}
