package rugs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestNHLCommand_Player(t *testing.T) {
	cmd := &NHLCommand{}
	ctx := NewTestContext()

	playerName := "McDavid"
	searchURL := fmt.Sprintf("https://search.d3.nhle.com/api/v1/search/player?culture=en-us&limit=5&q=%s", playerName)
	landingURL := "https://api-web.nhle.com/v1/player/8478402/landing"

	searchResponse := `[{
		"playerId": "8478402",
		"name": "Connor McDavid",
		"teamAbbrev": "EDM",
		"positionCode": "C"
	}]`

	landingResponse := `{
		"featuredStats": {
			"regularSeason": {
				"subSeason": {
					"season": 20232024,
					"goals": 32,
					"assists": 64,
					"points": 96,
					"plusMinus": 15
				}
			}
		},
		"position": "C",
		"headshot": "https://example.com/mcdavid.jpg"
	}`

	ctx.On("GetString", "type").Return("player")
	ctx.On("GetString", "name").Return(playerName)
	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, searchURL, map[string]string(nil)).Return(searchResponse, nil)
	ctx.On("HTTPGetString", 10, landingURL, map[string]string(nil)).Return(landingResponse, nil)
	ctx.On("FollowUpEmbed", mock.Anything).Return(nil)

	err := cmd.Execute(ctx)

	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}

func TestNHLCommand_Team(t *testing.T) {
	cmd := &NHLCommand{}
	ctx := NewTestContext()

	teamName := "Oilers"

	standingsResponse := `{
		"standings": [{
			"teamName": {"default": "Edmonton Oilers"},
			"teamAbbrev": {"default": "EDM"},
			"teamLogo": "https://example.com/edm.svg",
			"conferenceName": "Western",
			"divisionName": "Pacific",
			"gamesPlayed": 50,
			"wins": 30,
			"losses": 15,
			"otLosses": 5,
			"points": 65,
			"pointPctg": 0.650,
			"goalFor": 180,
			"goalAgainst": 150,
			"goalDifferential": 30,
			"streakCode": "W",
			"streakCount": 3
		}]
	}`

	ctx.On("GetString", "type").Return("team")
	ctx.On("GetString", "name").Return(teamName)
	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, mock.AnythingOfType("string"), map[string]string(nil)).Return(standingsResponse, nil)
	ctx.On("FollowUpEmbed", mock.Anything).Return(nil)

	err := cmd.Execute(ctx)

	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}

func TestNHLCommand_DefaultsToTeam(t *testing.T) {
	cmd := &NHLCommand{}
	ctx := NewTestContext()

	standingsResponse := `{
		"standings": [{
			"teamName": {"default": "Vancouver Canucks"},
			"teamAbbrev": {"default": "VAN"},
			"teamLogo": "https://example.com/van.svg",
			"conferenceName": "Western",
			"divisionName": "Pacific",
			"gamesPlayed": 50,
			"wins": 35,
			"losses": 10,
			"otLosses": 5,
			"points": 75,
			"pointPctg": 0.750,
			"goalFor": 200,
			"goalAgainst": 140,
			"goalDifferential": 60,
			"streakCode": "W",
			"streakCount": 5
		}]
	}`

	// Empty type should default to team
	ctx.On("GetString", "type").Return("")
	ctx.On("GetString", "name").Return("Canucks")
	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, mock.AnythingOfType("string"), map[string]string(nil)).Return(standingsResponse, nil)
	ctx.On("FollowUpEmbed", mock.Anything).Return(nil)

	err := cmd.Execute(ctx)

	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}
