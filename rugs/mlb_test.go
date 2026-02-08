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

	searchResponse := `{
		"search_player_all": {
			"queryResults": {
				"totalSize": "1",
				"row": {
					"player_id": "660271",
					"name_display_first_last": "Shohei Ohtani",
					"team_full": "Los Angeles Dodgers",
					"position": "DH"
				}
			}
		}
	}`

	statsResponse := `{
		"sport_hitting_tm": {
			"queryResults": {
				"totalSize": "1",
				"row": {
					"avg": ".304",
					"hr": "54",
					"rbi": "130",
					"ops": "1.036"
				}
			}
		}
	}`

	ctx.On("GetString", "type").Return("player")
	ctx.On("GetString", "name").Return(playerName)
	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, mock.AnythingOfType("string"), map[string]string(nil)).Return(searchResponse, nil).Once()
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
		"standings_schedule_date": {
			"standings_all": {
				"queryResults": {
					"totalSize": "1",
					"row": {
						"team_short": "NYY",
						"team_full": "New York Yankees",
						"w": "94",
						"l": "68",
						"pct": ".580",
						"gb": "0",
						"division": "AL East"
					}
				}
			}
		}
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
		"standings_schedule_date": {
			"standings_all": {
				"queryResults": {
					"totalSize": "1",
					"row": {
						"team_short": "LAD",
						"team_full": "Los Angeles Dodgers",
						"w": "98",
						"l": "64",
						"pct": ".605",
						"gb": "0",
						"division": "NL West"
					}
				}
			}
		}
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
