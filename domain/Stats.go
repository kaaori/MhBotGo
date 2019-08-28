package domain

// Stats : The model for the Stats table
type Stats struct {
	// StatsID             int32
	ServerID            string
	CommandsUsed        int32
	EventsRun           int32
	LastEventTimeUnix   int32
	BirthdaysRegistered int32

	// ORM Fields
	Server DiscordServer
}
