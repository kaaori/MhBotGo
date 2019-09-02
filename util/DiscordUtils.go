package util

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
	"github.com/kaaori/MhBotGo/domain"
)

var (
	// MhColor : The default embed colour
	MhColor = 0x9400d3

	// MhThumb : The default thumbnail image for embeds
	MhThumb = "https://i.imgur.com/erRDVM7.png"
)

func init() {
	log.Info("Initialising RPC Client")
}

// GetEmbed : Get Embed by parameters
func GetEmbed(title string, footer string, withThumb bool, fields ...*discordgo.MessageEmbedField) *discordgo.MessageEmbed {
	url := ""
	if withThumb {
		url = MhThumb
	}
	return &discordgo.MessageEmbed{
		// Author:      &discordgo.MessageEmbedAuthor{},
		Color: MhColor, // Green
		// Description: "This is a discordgo embed",
		Fields: fields,
		Thumbnail: &discordgo.MessageEmbedThumbnail{
			URL: url,
		},
		// Timestamp: time.Now().Format(time.RFC3339), // Discord wants ISO8601; RFC3339 is an extension of ISO8601 and should be completely compatible.
		Title:  title,
		Footer: &discordgo.MessageEmbedFooter{Text: footer}}
}

// GetField : Generate a field object more easily
func GetField(name string, text string, inline bool) *discordgo.MessageEmbedField {
	return &discordgo.MessageEmbedField{
		Value:  text,
		Name:   name,
		Inline: inline}
}

// SetBotGame : Sets the bot's status to playing a game when an event is passed, or just a game if not
func SetBotGame(s *discordgo.Session, game string, evt *domain.Event) {
	if evt != nil {
		go cycleEventParams(s, evt)
	} else {
		if err := s.UpdateStatus(0, game); err != nil {
			log.Error("Update status error: ", err)
			return
		}
	}
}

func cycleEventParams(s *discordgo.Session, evt *domain.Event) {
	i := 0

	for range time.NewTicker(4 * time.Second).C {
		// log.Info("Updating schedule")
		defer func() {
			if err := recover(); err != nil {
				// if we're in here, we had a panic and have caught it
				fmt.Printf("Panic deferred in scheduler: %s\n", err)
			}
		}()

		// If the event is over
		if time.Now().After(evt.EndTime) {
			SetBotGame(s, "<3", nil)
			break
		}

		switch i % 2 {
		case 0:
			SetBotGame(s, evt.EventName, nil)
			break
		case 1:
			SetBotGame(s, evt.EventLocation+" with "+evt.HostName, nil)
			break
		}
		i++
	}
}
