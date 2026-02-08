package rugs

import (
	"time"

	"github.com/bwmarrin/discordgo"
)

// EmbedBuilder provides a fluent interface for building Discord embeds
type EmbedBuilder struct {
	embed *discordgo.MessageEmbed
}

// NewEmbed creates a new embed builder
func NewEmbed() *EmbedBuilder {
	return &EmbedBuilder{
		embed: &discordgo.MessageEmbed{},
	}
}

// SetTitle sets the embed title
func (e *EmbedBuilder) SetTitle(title string) *EmbedBuilder {
	e.embed.Title = title
	return e
}

// SetDescription sets the embed description
func (e *EmbedBuilder) SetDescription(description string) *EmbedBuilder {
	e.embed.Description = description
	return e
}

// SetURL sets the embed URL
func (e *EmbedBuilder) SetURL(url string) *EmbedBuilder {
	e.embed.URL = url
	return e
}

// SetColor sets the embed color (decimal color value)
func (e *EmbedBuilder) SetColor(color int) *EmbedBuilder {
	e.embed.Color = color
	return e
}

// SetTimestamp sets the embed timestamp to now
func (e *EmbedBuilder) SetTimestamp() *EmbedBuilder {
	e.embed.Timestamp = time.Now().Format(time.RFC3339)
	return e
}

// SetFooter sets the embed footer
func (e *EmbedBuilder) SetFooter(text string, iconURL string) *EmbedBuilder {
	e.embed.Footer = &discordgo.MessageEmbedFooter{
		Text:    text,
		IconURL: iconURL,
	}
	return e
}

// SetThumbnail sets the embed thumbnail image
func (e *EmbedBuilder) SetThumbnail(url string) *EmbedBuilder {
	e.embed.Thumbnail = &discordgo.MessageEmbedThumbnail{
		URL: url,
	}
	return e
}

// SetImage sets the embed main image
func (e *EmbedBuilder) SetImage(url string) *EmbedBuilder {
	e.embed.Image = &discordgo.MessageEmbedImage{
		URL: url,
	}
	return e
}

// SetAuthor sets the embed author
func (e *EmbedBuilder) SetAuthor(name string, url string, iconURL string) *EmbedBuilder {
	e.embed.Author = &discordgo.MessageEmbedAuthor{
		Name:    name,
		URL:     url,
		IconURL: iconURL,
	}
	return e
}

// AddField adds a field to the embed
func (e *EmbedBuilder) AddField(name string, value string, inline bool) *EmbedBuilder {
	if e.embed.Fields == nil {
		e.embed.Fields = make([]*discordgo.MessageEmbedField, 0)
	}
	e.embed.Fields = append(e.embed.Fields, &discordgo.MessageEmbedField{
		Name:   name,
		Value:  value,
		Inline: inline,
	})
	return e
}

// Build returns the constructed embed
func (e *EmbedBuilder) Build() *discordgo.MessageEmbed {
	return e.embed
}

// Common color constants
const (
	ColorBlue    = 3447003
	ColorGreen   = 3066993
	ColorRed     = 15158332
	ColorOrange  = 15105570
	ColorPurple  = 10181046
	ColorGold    = 15844367
	ColorTeal    = 1752220
	ColorMagenta = 12320855
)
