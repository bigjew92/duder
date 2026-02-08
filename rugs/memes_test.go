package rugs

import (
	"encoding/json"
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestMemeCommandList(t *testing.T) {
	cmd := &MemeCommand{}
	ctx := NewTestContext()

	// Mock subcommand
	ctx.On("Interaction").Return(&discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				Options: []*discordgo.ApplicationCommandInteractionDataOption{
					{Name: "list"},
				},
			},
		},
	})

	// Mock Imgflip API for list
	templates := struct {
		Success bool `json:"success"`
		Data    struct {
			Memes []memeTemplate `json:"memes"`
		} `json:"data"`
	}{
		Success: true,
		Data: struct {
			Memes []memeTemplate `json:"memes"`
		}{
			Memes: []memeTemplate{
				{ID: "1", Name: "Drake Hotline Bling"},
				{ID: "2", Name: "Distracted Boyfriend"},
			},
		},
	}
	respInfo, _ := json.Marshal(templates)

	ctx.On("HTTPGetString", 10, "https://api.imgflip.com/get_memes", map[string]string(nil)).
		Return(string(respInfo), nil)

	ctx.On("Reply", mock.MatchedBy(func(s string) bool {
		return assert.Contains(t, s, "Drake Hotline Bling") && assert.Contains(t, s, "Distracted Boyfriend")
	})).Return(nil)

	err := cmd.Execute(ctx)
	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}
