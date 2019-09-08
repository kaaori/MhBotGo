package domain

// ScheduleView : The view for the template schedule html file
type ScheduleView struct {
	ServerName        string
	CurrentWeekString string
	Tz                string
	FactTitle         string
	Fact              string
	Week              *WeekView
}
