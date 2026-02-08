package rugs

import (
	"strings"
	"testing"

	"os"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestWeatherCommand(t *testing.T) {
	cmd := &WeatherCommand{}
	ctx := NewTestContext()

	// Mock top-level options
	ctx.On("GetString", "set_key").Return("")
	ctx.On("GetString", "set_default").Return("")
	ctx.On("GetString", "location").Return("Seattle, WA")

	// Verify ApplicationCommandData is NOT called for subcommands anymore
	// ctx.On("Interaction")... (Removed)

	// Mock storage for API key
	// We need to ensure getAPIKey returns a value.
	// Since we can't easily mock the storage interna without changing the test structure,
	// we'll rely on the default behavior or mock the storage if possible.
	// However, WeatherCommand creates its own storage if nil.
	// For testing, we might need to inject storage or set the env var.
	os.Setenv("OPENWEATHER_API_KEY", "test-key")
	defer os.Unsetenv("OPENWEATHER_API_KEY")

	// Mock HTTP request for geocoding
	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, mock.MatchedBy(func(url string) bool {
		return strings.Contains(url, "geo/1.0/direct") && strings.Contains(url, "Seattle")
	}), map[string]string(nil)).
		Return(`[{"name":"Seattle","lat":47.6062,"lon":-122.3321,"country":"US","state":"Washington"}]`, nil)

	// Mock HTTP request for forecast
	ctx.On("HTTPGetString", 10, mock.MatchedBy(func(url string) bool {
		return strings.Contains(url, "data/2.5/forecast") && strings.Contains(url, "lat=47.6062")
	}), map[string]string(nil)).
		Return(`{"list":[{"dt":1600000000,"main":{"temp":65.0,"feels_like":63.0,"humidity":50},"weather":[{"description":"clear sky"}],"wind":{"speed":5.0},"dt_txt":"2020-09-13 12:00:00"}]}`, nil)

	// Expect an embed reply
	ctx.On("FollowUpEmbed", mock.AnythingOfType("*discordgo.MessageEmbed")).Return(nil)

	err := cmd.Execute(ctx)
	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}
