// Modified template from https://github.com/2Bot/2Bot-Discord-Bot
package main

import (
	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/MhBotGo/util"
	"github.com/spf13/viper"
)

func readyEvent(s *discordgo.Session, ready *discordgo.Ready) {
	log.Trace("received ready event")
	util.SetBotGame(s, viper.GetString("game"), nil)
}

func guildJoinEvent(s *discordgo.Session, guild *discordgo.GuildCreate) {
	if server, err := BotInstance.ServerDao.GetServerByID(guild.ID); err != nil {
		log.Error("Error occured looking for guild!", err)
	} else if server == nil {
		log.Info("New guild detected, initialising database for " + guild.Name)
		initDbForGuild(guild)
	}
}
