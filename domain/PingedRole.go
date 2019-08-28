package domain

import (
	"github.com/bwmarrin/discordgo"
)

// PingedRole : A role to be pinged with an event
type PingedRole struct {
	PingedRoleID string
	EventID      int32

	// ORM Fields
	PingedRole  discordgo.Role
	LinkedEvent Event
}
