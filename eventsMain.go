// Modified template from https://github.com/2Bot/2Bot-Discord-Bot
package main

import (
	"github.com/bwmarrin/discordgo"
)

func readyEvent(s *discordgo.Session, m *discordgo.Ready) {
	log.Trace("received ready event")
	setBotGame(s)
}

func guildJoinEvent(s *discordgo.Session, m *discordgo.GuildCreate) {
	log.Trace("Joined server")

	// If not in DB
	log.Trace("Initialising database for " + m.Name)
	initDbForGuild(m)
	// endif
}
