package domain

// DayView : The list of events for a given day and if it is the current day
type DayView struct {
	Events             []*EventView
	Birthdays          []*BirthdayView
	IsCurrentDayString string
	DayName            string
}
