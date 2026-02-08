package rugs

import (
	"testing"

	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestCatFactCommand(t *testing.T) {
	cmd := &CatFactCommand{}
	ctx := NewTestContext()

	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, "https://catfact.ninja/fact", map[string]string(nil)).
		Return(`{"fact":"Cats are cool."}`, nil)
	ctx.On("FollowUp", "```Cats are cool.```").Return(nil)

	err := cmd.Execute(ctx)
	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}

func TestCatPicCommand(t *testing.T) {
	cmd := &CatPicCommand{}
	ctx := NewTestContext()

	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, "https://api.thecatapi.com/v1/images/search", map[string]string(nil)).
		Return(`[{"url":"https://example.com/cat.jpg"}]`, nil)

	ctx.On("FollowUpEmbed", mock.MatchedBy(func(embed *discordgo.MessageEmbed) bool {
		return embed.Image.URL == "https://example.com/cat.jpg"
	})).Return(nil)

	err := cmd.Execute(ctx)
	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}

func TestCatGifCommand(t *testing.T) {
	cmd := &CatGifCommand{}
	ctx := NewTestContext()

	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, "https://api.thecatapi.com/v1/images/search?mime_types=gif", map[string]string(nil)).
		Return(`[{"url":"https://example.com/cat.gif"}]`, nil)

	ctx.On("FollowUpEmbed", mock.MatchedBy(func(embed *discordgo.MessageEmbed) bool {
		return embed.Image.URL == "https://example.com/cat.gif"
	})).Return(nil)

	err := cmd.Execute(ctx)
	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}
