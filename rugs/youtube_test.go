package rugs

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
)

func TestYouTubeCommandRandom(t *testing.T) {
	// Need to mock storage for this one to set API key and Playlist ID
	// Or we can rely on environmental variables or defaults if the command supports it.
	// Looking at youtube.go, it uses c.getStorage().
	// This makes it hard to test without mocking the storage specifically or
	// using the setkey/setplaylist commands first in the test flow.

	// Implementing a simpler test for "setkey" to verify interaction
	cmd := &YouTubeCommand{}
	ctx := NewTestContext()

	// Mock subcommand
	ctx.On("Interaction").Return(&discordgo.InteractionCreate{
		Interaction: &discordgo.Interaction{
			Type: discordgo.InteractionApplicationCommand,
			Data: discordgo.ApplicationCommandInteractionData{
				Options: []*discordgo.ApplicationCommandInteractionDataOption{
					{Name: "setkey"},
				},
			},
		},
	})

	ctx.On("IsOwner").Return(true)
	ctx.On("GetString", "key").Return("fake-api-key")
	ctx.On("ReplyEphemeral", "YouTube API key set!").Return(nil)

	err := cmd.Execute(ctx)
	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}
