package rugs

import (
	"github.com/bwmarrin/discordgo"
	"github.com/stretchr/testify/mock"
)

// TestContext is a mock implementation of CommandContext for testing
type TestContext struct {
	mock.Mock
	SessionVal     *discordgo.Session
	InteractionVal *discordgo.InteractionCreate
	GuildVal       *discordgo.Guild
	ChannelVal     *discordgo.Channel
	UserVal        *discordgo.User
	OptionsVal     map[string]interface{} // Simplified for testing
}

func NewTestContext() *TestContext {
	return &TestContext{
		OptionsVal: make(map[string]interface{}),
	}
}

func (ctx *TestContext) Name() string {
	args := ctx.Called()
	return args.String(0)
}

func (ctx *TestContext) Session() *discordgo.Session {
	args := ctx.Called()
	return args.Get(0).(*discordgo.Session)
}

func (ctx *TestContext) Interaction() *discordgo.InteractionCreate {
	args := ctx.Called()
	return args.Get(0).(*discordgo.InteractionCreate)
}

func (ctx *TestContext) Guild() *discordgo.Guild {
	args := ctx.Called()
	return args.Get(0).(*discordgo.Guild)
}

func (ctx *TestContext) Channel() *discordgo.Channel {
	args := ctx.Called()
	return args.Get(0).(*discordgo.Channel)
}

func (ctx *TestContext) User() *discordgo.User {
	args := ctx.Called()
	return args.Get(0).(*discordgo.User)
}

func (ctx *TestContext) GetString(name string) string {
	args := ctx.Called(name)
	return args.String(0)
}

func (ctx *TestContext) GetBool(name string) bool {
	args := ctx.Called(name)
	return args.Bool(0)
}

func (ctx *TestContext) GetInt(name string) int64 {
	args := ctx.Called(name)
	return args.Get(0).(int64)
}

func (ctx *TestContext) GetFloat(name string) float64 {
	args := ctx.Called(name)
	return args.Get(0).(float64)
}

func (ctx *TestContext) GetUser(name string) *discordgo.User {
	args := ctx.Called(name)
	if args.Get(0) == nil {
		return nil
	}
	return args.Get(0).(*discordgo.User)
}

func (ctx *TestContext) Reply(content string) error {
	args := ctx.Called(content)
	return args.Error(0)
}

func (ctx *TestContext) ReplyEmbed(embed *discordgo.MessageEmbed) error {
	args := ctx.Called(embed)
	return args.Error(0)
}

func (ctx *TestContext) ReplyEphemeral(content string) error {
	args := ctx.Called(content)
	return args.Error(0)
}

func (ctx *TestContext) DeferReply() error {
	args := ctx.Called()
	return args.Error(0)
}

func (ctx *TestContext) FollowUp(content string) error {
	args := ctx.Called(content)
	return args.Error(0)
}

func (ctx *TestContext) FollowUpEmbed(embed *discordgo.MessageEmbed) error {
	args := ctx.Called(embed)
	return args.Error(0)
}

func (ctx *TestContext) HTTPGetString(timeout int, uri string, headers map[string]string) (string, error) {
	args := ctx.Called(timeout, uri, headers)
	return args.String(0), args.Error(1)
}

func (ctx *TestContext) HTTPPostString(timeout int, uri string, data map[string]string) (string, error) {
	args := ctx.Called(timeout, uri, data)
	return args.String(0), args.Error(1)
}

func (ctx *TestContext) IsOwner() bool {
	args := ctx.Called()
	return args.Bool(0)
}
