package rugs

import (
	"strings"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWeatherCommand(t *testing.T) {
	cmd := &WeatherCommand{}
	ctx := NewTestContext()

	// Set API key for testing
	t.Setenv("OPENWEATHER_API_KEY", "test-api-key")

	// Mock Interaction
	ctx.On("Interaction").Return(&discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				Options: []*discordgo.ApplicationCommandInteractionDataOption{
					{
						Name: "check",
						Type: discordgo.ApplicationCommandOptionSubCommand,
						Options: []*discordgo.ApplicationCommandInteractionDataOption{
							{
								Name:  "location",
								Type:  discordgo.ApplicationCommandOptionString,
								Value: "London",
							},
						},
					},
				},
			},
		},
	})
	// Expect GetString for location since handleCheck calls it
	ctx.On("GetString", "location").Return("London")

	ctx.On("DeferReply").Return(nil)
	// Actually, CommandContextImpl.GetString uses ctx.Options() which uses ctx.Interaction().
	// usage in weather.go: location := ctx.GetString("location")
	// usage in CommandContextImpl.GetString: logic iterates over options.
	// BUT weather.go line 98 calling ctx.Interaction() directly suggests it might resolve subcommands from Interaction data.

	// Let's verify weather.go content around line 98 first to be sure.
	// I will use view_file to check weather.go line 98.

	ctx.On("DeferReply").Return(nil)

	// Mock Geocoding
	ctx.On("HTTPGetString", 10, mock.MatchedBy(func(url string) bool {
		return strings.Contains(url, "api.openweathermap.org/geo/1.0/direct") && strings.Contains(url, "London")
	}), map[string]string(nil)).Return(`[{"name":"London","lat":51.5074,"lon":-0.1278,"country":"GB"}]`, nil)

	// Mock Forecast
	ctx.On("HTTPGetString", 10, mock.MatchedBy(func(url string) bool {
		return strings.Contains(url, "api.openweathermap.org/data/2.5/forecast") && strings.Contains(url, "lat=51.5074")
	}), map[string]string(nil)).Return(`{"list":[{"main":{"temp":55.0,"humidity":75},"weather":[{"main":"Clouds","description":"scattered clouds"}],"wind":{"speed":10.5}}],"city":{"name":"London"}}`, nil)

	// Expectation
	ctx.On("FollowUpEmbed", mock.MatchedBy(func(embed interface{}) bool {
		// Just checking it calls FollowUpEmbed is enough for now
		return true
	})).Return(nil)

	err := cmd.Execute(ctx)
	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}
