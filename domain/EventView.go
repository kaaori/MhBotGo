package domain

// EventView : The view object for the schedule HTML
type EventView struct {
	PrettyPrint    string
	StartTimestamp int64
	HasPassed      bool
	DayOfWeek      string
	HostName       string
	HostLocation   string
	EventName      string
}
