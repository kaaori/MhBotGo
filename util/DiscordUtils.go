package util

import (
	"log"

	"github.com/bwmarrin/discordgo"
)

var (
	// MhColor : The default embed colour
	MhColor = 0x9400d3

	// MhThumb : The default thumbnail image for embeds
	MhThumb      = "https://i.imgur.com/erRDVM7.png"
	eventRunning = false
)

func init() {
	log.Println("Initialising RPC Client")
}

// GetEmbed : Get Embed by parameters
func GetEmbed(title string, footer string, withThumb bool, customThumb string, fields ...*discordgo.MessageEmbedField) *discordgo.MessageEmbed {
	url := ""
	if withThumb && "" == customThumb {
		url = MhThumb
	} else if withThumb && customThumb != "" {
		url = customThumb
	}
	return &discordgo.MessageEmbed{
		// Author:      &discordgo.MessageEmbedAuthor{},
		Color: MhColor,
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
