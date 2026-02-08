package rugs

import (
	"encoding/json"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&CatFactCommand{})
	Register(&CatPicCommand{})
	Register(&CatGifCommand{})
}

// CatFactCommand implements the /catfact slash command
type CatFactCommand struct{}

func (c *CatFactCommand) Name() string {
	return "catfact"
}

func (c *CatFactCommand) Description() string {
	return "Get a random cat fact"
}

func (c *CatFactCommand) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

func (c *CatFactCommand) Execute(ctx CommandContext) error {
	ctx.DeferReply()

	resp, err := ctx.HTTPGetString(10, "https://catfact.ninja/fact", nil)
	if err != nil {
		return ctx.FollowUp("Failed to fetch cat fact. Try again later.")
	}

	var result struct {
		Fact string `json:"fact"`
	}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return ctx.FollowUp("Failed to parse cat fact.")
	}

	return ctx.FollowUp(fmt.Sprintf("```%s```", result.Fact))
}

// CatPicCommand implements the /catpic slash command
type CatPicCommand struct{}

func (c *CatPicCommand) Name() string {
	return "catpic"
}

func (c *CatPicCommand) Description() string {
	return "Get a random cat picture"
}

func (c *CatPicCommand) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

func (c *CatPicCommand) Execute(ctx CommandContext) error {
	ctx.DeferReply()

	resp, err := ctx.HTTPGetString(10, "https://api.thecatapi.com/v1/images/search", nil)
	if err != nil {
		return ctx.FollowUp("Failed to fetch cat picture. Try again later.")
	}

	var results []struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(resp), &results); err != nil || len(results) == 0 {
		return ctx.FollowUp("Failed to parse cat picture.")
	}

	embed := NewEmbed().
		SetTitle("Random Cat 🐱").
		SetImage(results[0].URL).
		SetColor(ColorOrange).
		Build()

	return ctx.FollowUpEmbed(embed)
}

// CatGifCommand implements the /catgif slash command
type CatGifCommand struct{}

func (c *CatGifCommand) Name() string {
	return "catgif"
}

func (c *CatGifCommand) Description() string {
	return "Get a random cat GIF"
}

func (c *CatGifCommand) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

func (c *CatGifCommand) Execute(ctx CommandContext) error {
	ctx.DeferReply()

	// The Cat API with gif filter
	resp, err := ctx.HTTPGetString(10, "https://api.thecatapi.com/v1/images/search?mime_types=gif", nil)
	if err != nil {
		return ctx.FollowUp("Failed to fetch cat GIF. Try again later.")
	}

	var results []struct {
		URL string `json:"url"`
	}
	if err := json.Unmarshal([]byte(resp), &results); err != nil || len(results) == 0 {
		return ctx.FollowUp("Failed to parse cat GIF.")
	}

	embed := NewEmbed().
		SetTitle("Random Cat GIF 🐱").
		SetImage(results[0].URL).
		SetColor(ColorOrange).
		Build()

	return ctx.FollowUpEmbed(embed)
}
