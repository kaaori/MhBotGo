package domain

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

// Fact : Object representation of json
type Fact struct {
	UserID       string
	FactContent  string
	LastUsedUnix int64

	LastUsedTime time.Time
	User         *discordgo.User
}
