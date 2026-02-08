package rugs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestBananaBallCommand_PlayerHitting(t *testing.T) {
	cmd := &BananaBallCommand{}
	ctx := NewTestContext()

	playerName := "Skole"

	playersResponse := `[{
		"first_name": "Jake",
		"last_name": "Skole",
		"jersey_number": 7,
		"primary_position": {"label": "First Base"},
		"image": "abc123",
		"games_played": 45,
		"at_bats": 156,
		"batting_average": 0.312,
		"home_runs": 8,
		"runs_batted_in": 32,
		"on_base_plus_slugging": 0.895,
		"ball_four_sprints": 24
	}]`

	ctx.On("GetString", "type").Return("player")
	ctx.On("GetString", "name").Return(playerName)
	ctx.On("GetString", "stats").Return("")
	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, "https://stats.bananaball.com/stats/players_stats?category=hitting", map[string]string(nil)).Return(playersResponse, nil)
	ctx.On("FollowUpEmbed", mock.Anything).Return(nil)

	err := cmd.Execute(ctx)

	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}

func TestBananaBallCommand_PlayerPitching(t *testing.T) {
	cmd := &BananaBallCommand{}
	ctx := NewTestContext()

	playerName := "Luigs"

	playersResponse := `[{
		"first_name": "Kyle",
		"last_name": "Luigs",
		"jersey_number": 22,
		"primary_position": {"label": "Pitcher"},
		"image": "def456",
		"games_played": 20,
		"innings_pitched": 45.2,
		"earned_run_average": 2.56,
		"wins": 6,
		"losses": 2,
		"pitcher_strikeouts": 52,
		"ship": 1.45
	}]`

	ctx.On("GetString", "type").Return("player")
	ctx.On("GetString", "name").Return(playerName)
	ctx.On("GetString", "stats").Return("pitching")
	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, "https://stats.bananaball.com/stats/players_stats?category=pitching", map[string]string(nil)).Return(playersResponse, nil)
	ctx.On("FollowUpEmbed", mock.Anything).Return(nil)

	err := cmd.Execute(ctx)

	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}

func TestBananaBallCommand_Team(t *testing.T) {
	cmd := &BananaBallCommand{}
	ctx := NewTestContext()

	teamName := "Party Animals"

	teamsResponse := `[
		{"id": "1", "name": "Savannah Bananas", "abbreviation": "SAV", "record": {"wins": 42, "losses": 12}, "batting_average": 0.285, "earned_run_average": 3.21, "home_runs": 67, "ball_four_sprints": 156},
		{"id": "2", "name": "Party Animals", "abbreviation": "PA", "record": {"wins": 38, "losses": 16}, "batting_average": 0.272, "earned_run_average": 3.45, "home_runs": 54, "ball_four_sprints": 142},
		{"id": "3", "name": "Firefighters", "abbreviation": "FF", "record": {"wins": 35, "losses": 19}, "batting_average": 0.265, "earned_run_average": 3.67, "home_runs": 48, "ball_four_sprints": 128},
		{"id": "4", "name": "Texas Tailgaters", "abbreviation": "TT", "record": {"wins": 28, "losses": 26}, "batting_average": 0.258, "earned_run_average": 4.12, "home_runs": 41, "ball_four_sprints": 112}
	]`

	ctx.On("GetString", "type").Return("team")
	ctx.On("GetString", "name").Return(teamName)
	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, "https://stats.bananaball.com/stats/teams_stats?category=teams", map[string]string(nil)).Return(teamsResponse, nil)
	ctx.On("FollowUpEmbed", mock.Anything).Return(nil)

	err := cmd.Execute(ctx)

	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}

func TestBananaBallCommand_PlayerNotFound(t *testing.T) {
	cmd := &BananaBallCommand{}
	ctx := NewTestContext()

	playersResponse := `[]`

	ctx.On("GetString", "type").Return("player")
	ctx.On("GetString", "name").Return("NonexistentPlayer")
	ctx.On("GetString", "stats").Return("")
	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, "https://stats.bananaball.com/stats/players_stats?category=hitting", map[string]string(nil)).Return(playersResponse, nil)
	ctx.On("FollowUp", "No Banana Ball player found matching 'NonexistentPlayer'.").Return(nil)

	err := cmd.Execute(ctx)

	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}

func TestBananaBallCommand_DefaultsToPlayer(t *testing.T) {
	cmd := &BananaBallCommand{}
	ctx := NewTestContext()

	playersResponse := `[{
		"first_name": "Test",
		"last_name": "Player",
		"jersey_number": 1,
		"primary_position": {"label": "Outfield"},
		"games_played": 10,
		"at_bats": 30,
		"batting_average": 0.300,
		"home_runs": 2,
		"runs_batted_in": 8,
		"on_base_plus_slugging": 0.850,
		"ball_four_sprints": 5
	}]`

	// Empty type should default to player
	ctx.On("GetString", "type").Return("")
	ctx.On("GetString", "name").Return("Test")
	ctx.On("GetString", "stats").Return("")
	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, "https://stats.bananaball.com/stats/players_stats?category=hitting", map[string]string(nil)).Return(playersResponse, nil)
	ctx.On("FollowUpEmbed", mock.Anything).Return(nil)

	err := cmd.Execute(ctx)

	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}
