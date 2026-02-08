package rugs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&StarWarsCommand{})
}

// StarWarsCommand implements the /starwars slash command
type StarWarsCommand struct{}

const starWarsAPIBase = "https://starwars-databank-server.vercel.app/api/v1"

func (c *StarWarsCommand) Name() string {
	return "starwars"
}

func (c *StarWarsCommand) Description() string {
	return "Search the Star Wars databank"
}

func (c *StarWarsCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "type",
			Description: "What to search for",
			Required:    true,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "Character", Value: "characters"},
				{Name: "Planet", Value: "locations"},
				{Name: "Vehicle/Starship", Value: "vehicles"},
				{Name: "Species", Value: "species"},
				{Name: "Droid", Value: "droids"},
				{Name: "Organization", Value: "organizations"},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "name",
			Description: "Name to search for",
			Required:    true,
		},
	}
}

func (c *StarWarsCommand) Execute(ctx CommandContext) error {
	searchType := ctx.GetString("type")
	name := ctx.GetString("name")

	ctx.DeferReply()

	// Fetch all items for the category (limit=1000 should cover everything)
	url := fmt.Sprintf("%s/%s?page=1&limit=1000", starWarsAPIBase, searchType)
	resp, err := ctx.HTTPGetString(10, url, nil)
	if err != nil {
		return ctx.FollowUp("Failed to access Star Wars databank.")
	}

	// Generic structure for all Databank items
	type DatabankItem struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		Image       string `json:"image"`
	}

	var result struct {
		Data []DatabankItem `json:"data"`
	}

	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return ctx.FollowUp("Failed to parse databank response.")
	}

	// Fuzzy search
	var match *DatabankItem
	nameLower := strings.ToLower(name)

	// 1. Exact match
	for _, item := range result.Data {
		if strings.ToLower(item.Name) == nameLower {
			match = &item
			break
		}
	}

	// 2. Contains match (if no exact match)
	if match == nil {
		for _, item := range result.Data {
			if strings.Contains(strings.ToLower(item.Name), nameLower) {
				match = &item
				break
			}
		}
	}

	if match == nil {
		return ctx.FollowUp(fmt.Sprintf("No results found for '%s' in %s.", name, searchType))
	}

	embed := NewEmbed().
		SetTitle(match.Name).
		SetDescription(match.Description).
		SetImage(match.Image).
		SetColor(ColorGold).
		Build()

	return ctx.FollowUpEmbed(embed)
}

// Remove unused handlers
