package util

import (
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
