package rugs

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&StarWarsCommand{})
}

// StarWarsCommand implements the /starwars slash command
type StarWarsCommand struct{}

const starWarsAPIBase = "https://swapi.dev/api"

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
				{Name: "Character", Value: "people"},
				{Name: "Planet", Value: "planets"},
				{Name: "Starship", Value: "starships"},
				{Name: "Vehicle", Value: "vehicles"},
				{Name: "Species", Value: "species"},
				{Name: "Film", Value: "films"},
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

	searchURL := fmt.Sprintf("%s/%s/?search=%s", starWarsAPIBase, searchType, url.QueryEscape(name))
	resp, err := ctx.HTTPGetString(10, searchURL, nil)
	if err != nil {
		return ctx.FollowUp("Failed to search Star Wars databank.")
	}

	switch searchType {
	case "people":
		return c.handlePerson(ctx, resp)
	case "planets":
		return c.handlePlanet(ctx, resp)
	case "starships":
		return c.handleStarship(ctx, resp)
	case "vehicles":
		return c.handleVehicle(ctx, resp)
	case "species":
		return c.handleSpecies(ctx, resp)
	case "films":
		return c.handleFilm(ctx, resp)
	}

	return ctx.FollowUp("Unknown search type.")
}

func (c *StarWarsCommand) handlePerson(ctx CommandContext, resp string) error {
	var result struct {
		Results []struct {
			Name      string `json:"name"`
			Height    string `json:"height"`
			Mass      string `json:"mass"`
			HairColor string `json:"hair_color"`
			EyeColor  string `json:"eye_color"`
			BirthYear string `json:"birth_year"`
			Gender    string `json:"gender"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(resp), &result); err != nil || len(result.Results) == 0 {
		return ctx.FollowUp("No characters found.")
	}

	p := result.Results[0]
	embed := NewEmbed().
		SetTitle(fmt.Sprintf("⚔️ %s", p.Name)).
		AddField("Height", p.Height+" cm", true).
		AddField("Mass", p.Mass+" kg", true).
		AddField("Birth Year", p.BirthYear, true).
		AddField("Gender", p.Gender, true).
		AddField("Hair Color", p.HairColor, true).
		AddField("Eye Color", p.EyeColor, true).
		SetColor(ColorGold).
		Build()

	return ctx.FollowUpEmbed(embed)
}

func (c *StarWarsCommand) handlePlanet(ctx CommandContext, resp string) error {
	var result struct {
		Results []struct {
			Name           string `json:"name"`
			Climate        string `json:"climate"`
			Terrain        string `json:"terrain"`
			Population     string `json:"population"`
			Diameter       string `json:"diameter"`
			OrbitalPeriod  string `json:"orbital_period"`
			RotationPeriod string `json:"rotation_period"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(resp), &result); err != nil || len(result.Results) == 0 {
		return ctx.FollowUp("No planets found.")
	}

	p := result.Results[0]
	embed := NewEmbed().
		SetTitle(fmt.Sprintf("🌍 %s", p.Name)).
		AddField("Climate", p.Climate, true).
		AddField("Terrain", p.Terrain, true).
		AddField("Population", p.Population, true).
		AddField("Diameter", p.Diameter+" km", true).
		AddField("Orbital Period", p.OrbitalPeriod+" days", true).
		AddField("Rotation Period", p.RotationPeriod+" hours", true).
		SetColor(ColorBlue).
		Build()

	return ctx.FollowUpEmbed(embed)
}

func (c *StarWarsCommand) handleStarship(ctx CommandContext, resp string) error {
	var result struct {
		Results []struct {
			Name             string `json:"name"`
			Model            string `json:"model"`
			Manufacturer     string `json:"manufacturer"`
			Class            string `json:"starship_class"`
			Length           string `json:"length"`
			MaxSpeed         string `json:"max_atmosphering_speed"`
			HyperdriveRating string `json:"hyperdrive_rating"`
			Crew             string `json:"crew"`
			Passengers       string `json:"passengers"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(resp), &result); err != nil || len(result.Results) == 0 {
		return ctx.FollowUp("No starships found.")
	}

	s := result.Results[0]
	embed := NewEmbed().
		SetTitle(fmt.Sprintf("🚀 %s", s.Name)).
		AddField("Model", s.Model, true).
		AddField("Class", s.Class, true).
		AddField("Manufacturer", s.Manufacturer, false).
		AddField("Length", s.Length+" m", true).
		AddField("Max Speed", s.MaxSpeed, true).
		AddField("Hyperdrive", s.HyperdriveRating, true).
		AddField("Crew", s.Crew, true).
		AddField("Passengers", s.Passengers, true).
		SetColor(ColorPurple).
		Build()

	return ctx.FollowUpEmbed(embed)
}

func (c *StarWarsCommand) handleVehicle(ctx CommandContext, resp string) error {
	var result struct {
		Results []struct {
			Name         string `json:"name"`
			Model        string `json:"model"`
			Manufacturer string `json:"manufacturer"`
			Class        string `json:"vehicle_class"`
			Length       string `json:"length"`
			MaxSpeed     string `json:"max_atmosphering_speed"`
			Crew         string `json:"crew"`
			Passengers   string `json:"passengers"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(resp), &result); err != nil || len(result.Results) == 0 {
		return ctx.FollowUp("No vehicles found.")
	}

	v := result.Results[0]
	embed := NewEmbed().
		SetTitle(fmt.Sprintf("🚗 %s", v.Name)).
		AddField("Model", v.Model, true).
		AddField("Class", v.Class, true).
		AddField("Manufacturer", v.Manufacturer, false).
		AddField("Length", v.Length+" m", true).
		AddField("Max Speed", v.MaxSpeed, true).
		AddField("Crew", v.Crew, true).
		AddField("Passengers", v.Passengers, true).
		SetColor(ColorOrange).
		Build()

	return ctx.FollowUpEmbed(embed)
}

func (c *StarWarsCommand) handleSpecies(ctx CommandContext, resp string) error {
	var result struct {
		Results []struct {
			Name            string `json:"name"`
			Classification  string `json:"classification"`
			Designation     string `json:"designation"`
			AverageHeight   string `json:"average_height"`
			SkinColors      string `json:"skin_colors"`
			HairColors      string `json:"hair_colors"`
			EyeColors       string `json:"eye_colors"`
			AverageLifespan string `json:"average_lifespan"`
			Language        string `json:"language"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(resp), &result); err != nil || len(result.Results) == 0 {
		return ctx.FollowUp("No species found.")
	}

	s := result.Results[0]
	embed := NewEmbed().
		SetTitle(fmt.Sprintf("👽 %s", s.Name)).
		AddField("Classification", s.Classification, true).
		AddField("Designation", s.Designation, true).
		AddField("Average Height", s.AverageHeight+" cm", true).
		AddField("Average Lifespan", s.AverageLifespan+" years", true).
		AddField("Language", s.Language, true).
		SetColor(ColorGreen).
		Build()

	return ctx.FollowUpEmbed(embed)
}

func (c *StarWarsCommand) handleFilm(ctx CommandContext, resp string) error {
	var result struct {
		Results []struct {
			Title        string `json:"title"`
			EpisodeID    int    `json:"episode_id"`
			OpeningCrawl string `json:"opening_crawl"`
			Director     string `json:"director"`
			Producer     string `json:"producer"`
			ReleaseDate  string `json:"release_date"`
		} `json:"results"`
	}
	if err := json.Unmarshal([]byte(resp), &result); err != nil || len(result.Results) == 0 {
		return ctx.FollowUp("No films found.")
	}

	f := result.Results[0]

	// Truncate opening crawl
	crawl := f.OpeningCrawl
	if len(crawl) > 200 {
		crawl = crawl[:200] + "..."
	}

	embed := NewEmbed().
		SetTitle(fmt.Sprintf("🎬 Episode %d: %s", f.EpisodeID, f.Title)).
		SetDescription(crawl).
		AddField("Director", f.Director, true).
		AddField("Producer", f.Producer, true).
		AddField("Release Date", f.ReleaseDate, true).
		SetColor(ColorGold).
		Build()

	return ctx.FollowUpEmbed(embed)
}
