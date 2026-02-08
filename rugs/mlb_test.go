package rugs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMLBCommand_Player(t *testing.T) {
	cmd := &MLBCommand{}
	ctx := NewTestContext()

	playerName := "Ohtani"

	playersResponse := `{
		"people": [{
			"id": 660271,
			"fullName": "Shohei Ohtani",
			"currentTeam": {"id": 119},
			"primaryPosition": {"abbreviation": "DH"}
		}]
	}`

	statsResponse := `{
		"stats": [{
			"splits": [{
				"stat": {
					"avg": ".304",
					"homeRuns": 54,
					"rbi": 130,
					"ops": "1.036"
				},
				"team": {"name": "Los Angeles Dodgers"}
			}]
		}]
	}`

	ctx.On("GetString", "type").Return("player")
	ctx.On("GetString", "name").Return(playerName)
	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, mock.AnythingOfType("string"), map[string]string(nil)).Return(playersResponse, nil).Once()
	ctx.On("HTTPGetString", 10, mock.AnythingOfType("string"), map[string]string(nil)).Return(statsResponse, nil).Once()
	ctx.On("FollowUpEmbed", mock.Anything).Return(nil)

	err := cmd.Execute(ctx)

	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}

func TestMLBCommand_Team(t *testing.T) {
	cmd := &MLBCommand{}
	ctx := NewTestContext()

	teamName := "Yankees"

	standingsResponse := `{
		"records": [{
			"division": {"id": 201, "name": "American League East"},
			"teamRecords": [{
				"team": {"id": 147, "name": "New York Yankees"},
				"wins": 94,
				"losses": 68,
				"winningPercentage": ".580",
				"gamesBack": "-",
				"divisionRank": "1"
			}]
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

func TestMLBCommand_DefaultsToTeam(t *testing.T) {
	cmd := &MLBCommand{}
	ctx := NewTestContext()

	standingsResponse := `{
		"records": [{
			"division": {"id": 203, "name": "National League West"},
			"teamRecords": [{
				"team": {"id": 119, "name": "Los Angeles Dodgers"},
				"wins": 98,
				"losses": 64,
				"winningPercentage": ".605",
				"gamesBack": "-",
				"divisionRank": "1"
			}]
		}]
	}`

	// Empty type should default to team
	ctx.On("GetString", "type").Return("")
	ctx.On("GetString", "name").Return("Dodgers")
	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, mock.AnythingOfType("string"), map[string]string(nil)).Return(standingsResponse, nil)
	ctx.On("FollowUpEmbed", mock.Anything).Return(nil)

	err := cmd.Execute(ctx)

	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}
