package rugs

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&WeatherCommand{})
}

// WeatherCommand implements the /weather slash command
type WeatherCommand struct {
	storage *Storage
}

const (
	geocodeURL  = "https://api.openweathermap.org/geo/1.0/direct"
	forecastURL = "https://api.openweathermap.org/data/2.5/forecast"
)

func (c *WeatherCommand) Name() string {
	return "weather"
}

func (c *WeatherCommand) Description() string {
	return "Get weather forecast for a location"
}

func (c *WeatherCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "location",
			Description: "Location to check (e.g., 'Seattle, WA')",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "set_default",
			Description: "Set your default location",
			Required:    false,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "set_key",
			Description: "Set the OpenWeatherMap API key (owner only)",
			Required:    false,
		},
	}
}

func (c *WeatherCommand) getStorage() *Storage {
	if c.storage == nil {
		c.storage = NewStorage("weather")
		c.storage.Load()
	}
	return c.storage
}

func (c *WeatherCommand) getAPIKey() string {
	storage := c.getStorage()
	if key, ok := storage.GetNested("settings", "api_key"); ok {
		if str, ok := key.(string); ok && str != "" {
			return str
		}
	}
	return os.Getenv("OPENWEATHER_API_KEY")
}

func (c *WeatherCommand) Execute(ctx CommandContext) error {
	// Handle set_key
	if key := ctx.GetString("set_key"); key != "" {
		return c.handleSetKey(ctx, key)
	}

	// Handle set_default
	if loc := ctx.GetString("set_default"); loc != "" {
		return c.handleSetLocation(ctx, loc)
	}

	// Handle weather check
	location := ctx.GetString("location")
	return c.handleCheck(ctx, location)
}

func (c *WeatherCommand) handleSetKey(ctx CommandContext, key string) error {
	if !ctx.IsOwner() {
		return ctx.ReplyEphemeral("Only the bot owner can set the API key.")
	}

	storage := c.getStorage()
	storage.SetNested(key, "settings", "api_key")
	storage.Save()

	return ctx.ReplyEphemeral("API key set successfully!")
}

func (c *WeatherCommand) handleSetLocation(ctx CommandContext, location string) error {
	storage := c.getStorage()
	storage.SetNested(location, "users", ctx.User().ID, "location")
	storage.Save()

	return ctx.Reply(fmt.Sprintf("Default location set to: %s", location))
}

func (c *WeatherCommand) handleCheck(ctx CommandContext, location string) error {
	apiKey := c.getAPIKey()
	if apiKey == "" {
		return ctx.ReplyEphemeral("Weather API key not configured. Ask the bot owner to set it with `/weather set_key:`.")
	}
	if location == "" {
		// Try to get saved location
		storage := c.getStorage()
		if saved, ok := storage.GetNested("users", ctx.User().ID, "location"); ok {
			if str, ok := saved.(string); ok {
				location = str
			}
		}
	}

	if location == "" {
		return ctx.ReplyEphemeral("Please provide a location or set a default with `/weather set_default:`.")
	}

	ctx.DeferReply()

	// Geocode the location
	geoURL := fmt.Sprintf("%s?q=%s&limit=1&appid=%s", geocodeURL, url.QueryEscape(location), apiKey)
	resp, err := ctx.HTTPGetString(10, geoURL, nil)
	if err != nil {
		return ctx.FollowUp(fmt.Sprintf("Failed to geocode location: %v", err))
	}

	var geoResults []struct {
		Name    string  `json:"name"`
		Country string  `json:"country"`
		State   string  `json:"state"`
		Lat     float64 `json:"lat"`
		Lon     float64 `json:"lon"`
	}
	if err := json.Unmarshal([]byte(resp), &geoResults); err != nil || len(geoResults) == 0 {
		return ctx.FollowUp("Location not found.")
	}

	geo := geoResults[0]

	// Get forecast
	fcURL := fmt.Sprintf("%s?lat=%f&lon=%f&units=imperial&appid=%s", forecastURL, geo.Lat, geo.Lon, apiKey)
	resp, err = ctx.HTTPGetString(10, fcURL, nil)
	if err != nil {
		return ctx.FollowUp("Failed to get forecast.")
	}

	var forecast struct {
		List []struct {
			Dt   int64 `json:"dt"`
			Main struct {
				Temp      float64 `json:"temp"`
				FeelsLike float64 `json:"feels_like"`
				Humidity  int     `json:"humidity"`
			} `json:"main"`
			Weather []struct {
				Description string `json:"description"`
				Icon        string `json:"icon"`
			} `json:"weather"`
			Wind struct {
				Speed float64 `json:"speed"`
			} `json:"wind"`
			DtTxt string `json:"dt_txt"`
		} `json:"list"`
	}
	if err := json.Unmarshal([]byte(resp), &forecast); err != nil {
		return ctx.FollowUp("Failed to parse forecast.")
	}

	// Build embed with next 3 forecasts
	locationStr := geo.Name
	if geo.State != "" {
		locationStr += ", " + geo.State
	}
	locationStr += ", " + geo.Country

	embed := NewEmbed().
		SetTitle(fmt.Sprintf("🌤️ Weather for %s", locationStr)).
		SetColor(ColorBlue)

	// Show next 3 time periods
	for i := 0; i < 3 && i < len(forecast.List); i++ {
		f := forecast.List[i]
		desc := ""
		if len(f.Weather) > 0 {
			desc = strings.Title(f.Weather[0].Description)
		}

		embed.AddField(
			f.DtTxt,
			fmt.Sprintf("🌡️ %.0f°F (feels like %.0f°F)\n%s\n💨 %.0f mph | 💧 %d%%",
				f.Main.Temp, f.Main.FeelsLike, desc, f.Wind.Speed, f.Main.Humidity),
			false,
		)
	}

	return ctx.FollowUpEmbed(embed.Build())
}
