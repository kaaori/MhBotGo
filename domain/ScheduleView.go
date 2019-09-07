package domain

// ScheduleView : The view for the template schedule html file
type ScheduleView struct {
	ServerName        string
	CurrentWeekString string
	Tz                string
	FactTitle         string
	Fact              string
	MondayEvents      []*EventView
	TuesdayEvents     []*EventView
	WednesdayEvents   []*EventView
	ThursdayEvents    []*EventView
	FridayEvents      []*EventView
	SaturdayEvents    []*EventView
	SundayEvents      []*EventView
}
