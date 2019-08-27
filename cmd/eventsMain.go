// Modified template from https://github.com/2Bot/2Bot-Discord-Bot
package main

import (
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
)

func messageCreateEvent(s *discordgo.Session, m *discordgo.MessageCreate) {
	if m.Author.Bot {
		return
	}

	if m.Author.ID == s.State.User.ID {
		return
	}

	// if m.Content == "Test!" {
	// 	s.ChannelMessageSend(m.ChannelID, "Toast!")
	// }

	// if m.Content == "Toast!" {
	// 	s.ChannelMessageSend(m.ChannelID, "Test!")
	// }

	guildDetails, err := guildDetails(m.ChannelID, "", s)
	if err != nil {
		return
	}

	prefix, err := activePrefix(m.ChannelID, s)
	if err != nil {
		return
	}

	if !strings.HasPrefix(m.Content, viper.GetString("prefix")) && !strings.HasPrefix(m.Content, prefix) {
		return
	}

	parseCommand(s, m, guildDetails, func() string {
		if strings.HasPrefix(m.Content, viper.GetString("prefix")) {
			fmt.Println("Prefix is " + viper.GetString("prefix"))
			return strings.TrimPrefix(m.Content, viper.GetString("prefix"))
		}
		return strings.TrimPrefix(m.Content, prefix)
	}())
}

func readyEvent(s *discordgo.Session, m *discordgo.Ready) {
	log.Trace("received ready event")
	setBotGame(s)
}

func guildJoinEvent(s *discordgo.Session, m *discordgo.Ready) {
	log.Trace("Joined server")
}
