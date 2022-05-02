// Modified template from https://github.com/2Bot/2Bot-Discord-Bot
package main

import (
	"log"

	"github.com/bwmarrin/discordgo"
	"github.com/spf13/viper"
	"mhbotgo.com/dao"
	"mhbotgo.com/scheduler"
)

func readyEvent(s *discordgo.Session, ready *discordgo.Ready) {
	log.Println("received ready event")
	BotInstance.SetBotGame(s, viper.GetString("game"))
}

func guildJoinEvent(s *discordgo.Session, guild *discordgo.GuildCreate) {
	scheduler.GuildsWithNoEvents = append(scheduler.GuildsWithNoEvents, guild.ID)

	DB := dao.GetConnection(dao.ConnString)
	defer DB.Close()

	if DB == nil {
		panic("DB was nil")
	}
	if server, err := BotInstance.ServerDao.GetServerByID(guild.ID); err != nil {
		log.Fatal("Error occured looking for guild!", err)
	} else if err != nil {
		log.Fatal("Error getting server", err)
	} else if server == nil {
		log.Println("New guild detected, initialising database for " + guild.Name)
		initDbForGuild(guild)
	} else if server != nil && server.Guild != nil {
		log.Println("Guild found: " + server.Guild.Name + ", love you<3")
		log.Println("\t\tChannels found: ", len(guild.Channels))
	}
}
