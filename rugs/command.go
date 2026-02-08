package rugs

import (
	"github.com/bwmarrin/discordgo"
)

// SlashCommand defines the interface for all slash commands
type SlashCommand interface {
	// Name returns the command name (e.g., "ping", "weather")
	Name() string

	// Description returns a short description shown in Discord's command list
	Description() string

	// Options returns the command options (arguments) for this command
	Options() []*discordgo.ApplicationCommandOption

	// Execute handles the command interaction
	Execute(ctx CommandContext) error
}

// EventHandler defines optional event handlers a command can implement
type EventHandler interface {
	// OnMessage is called for every message (for commands like lastseen)
	OnMessage(ctx *MessageContext)
}

// ReactionHandler defines reaction event handlers
type ReactionHandler interface {
	// OnReactionAdd is called when a reaction is added
	OnReactionAdd(ctx *ReactionContext)

	// OnReactionRemove is called when a reaction is removed
	OnReactionRemove(ctx *ReactionContext)
}

// PresenceHandler defines presence event handlers
type PresenceHandler interface {
	// OnPresenceUpdate is called when a user's presence changes
	OnPresenceUpdate(ctx *PresenceContext)
}

// CommandContext defines the interface for executing a slash command
type CommandContext interface {
	Name() string
	Session() *discordgo.Session
	Interaction() *discordgo.InteractionCreate
	Guild() *discordgo.Guild
	Channel() *discordgo.Channel
	User() *discordgo.User

	GetString(name string) string
	GetInt(name string) int64
	GetBool(name string) bool
	GetUser(name string) *discordgo.User

	Reply(content string) error
	ReplyEmbed(embed *discordgo.MessageEmbed) error
	ReplyEphemeral(content string) error
	DeferReply() error
	FollowUp(content string) error
	FollowUpEmbed(embed *discordgo.MessageEmbed) error
	HTTPGetString(timeout int, uri string, headers map[string]string) (string, error)
	HTTPPostString(timeout int, uri string, data map[string]string) (string, error)

	IsOwner() bool
}

// CommandContextImpl contains all information for executing a slash command
type CommandContextImpl struct {
	SessionVal     *discordgo.Session
	InteractionVal *discordgo.InteractionCreate
	GuildVal       *discordgo.Guild
	ChannelVal     *discordgo.Channel
	UserVal        *discordgo.User
	OptionsVal     map[string]*discordgo.ApplicationCommandInteractionDataOption
}

func (ctx *CommandContextImpl) Name() string {
	return ctx.InteractionVal.ApplicationCommandData().Name
}

func (ctx *CommandContextImpl) Session() *discordgo.Session {
	return ctx.SessionVal
}

func (ctx *CommandContextImpl) Interaction() *discordgo.InteractionCreate {
	return ctx.InteractionVal
}

func (ctx *CommandContextImpl) Guild() *discordgo.Guild {
	return ctx.GuildVal
}

func (ctx *CommandContextImpl) Channel() *discordgo.Channel {
	return ctx.ChannelVal
}

func (ctx *CommandContextImpl) User() *discordgo.User {
	return ctx.UserVal
}

// GetString returns a string option value
func (ctx *CommandContextImpl) GetString(name string) string {
	if opt, ok := ctx.OptionsVal[name]; ok {
		return opt.StringValue()
	}
	return ""
}

// GetInt returns an integer option value
func (ctx *CommandContextImpl) GetInt(name string) int64 {
	if opt, ok := ctx.OptionsVal[name]; ok {
		return opt.IntValue()
	}
	return 0
}

// GetBool returns a boolean option value
func (ctx *CommandContextImpl) GetBool(name string) bool {
	if opt, ok := ctx.OptionsVal[name]; ok {
		return opt.BoolValue()
	}
	return false
}

// GetUser returns a user option value
func (ctx *CommandContextImpl) GetUser(name string) *discordgo.User {
	if opt, ok := ctx.OptionsVal[name]; ok {
		return opt.UserValue(ctx.SessionVal)
	}
	return nil
}

// Reply sends a text response to the interaction
func (ctx *CommandContextImpl) Reply(content string) error {
	return ctx.SessionVal.InteractionRespond(ctx.InteractionVal.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

// ReplyEmbed sends an embed response to the interaction
func (ctx *CommandContextImpl) ReplyEmbed(embed *discordgo.MessageEmbed) error {
	return ctx.SessionVal.InteractionRespond(ctx.InteractionVal.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// ReplyEphemeral sends a private response only visible to the user
func (ctx *CommandContextImpl) ReplyEphemeral(content string) error {
	return ctx.SessionVal.InteractionRespond(ctx.InteractionVal.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// DeferReply acknowledges the interaction and allows for a delayed response
func (ctx *CommandContextImpl) DeferReply() error {
	return ctx.SessionVal.InteractionRespond(ctx.InteractionVal.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

// FollowUp sends a follow-up message after DeferReply
func (ctx *CommandContextImpl) FollowUp(content string) error {
	_, err := ctx.SessionVal.FollowupMessageCreate(ctx.InteractionVal.Interaction, true, &discordgo.WebhookParams{
		Content: content,
	})
	return err
}

func (ctx *CommandContextImpl) FollowUpEmbed(embed *discordgo.MessageEmbed) error {
	_, err := ctx.SessionVal.FollowupMessageCreate(ctx.InteractionVal.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
	})
	return err
}

func (ctx *CommandContextImpl) HTTPGetString(timeout int, uri string, headers map[string]string) (string, error) {
	return HTTPGetString(timeout, uri, headers)
}

func (ctx *CommandContextImpl) HTTPPostString(timeout int, uri string, data map[string]string) (string, error) {
	return HTTPPostString(timeout, uri, data)
}

// IsOwner returns true if the user is the bot owner
func (ctx *CommandContextImpl) IsOwner() bool {
	return ctx.UserVal.ID == ownerID
}

// MessageContext contains information for message events
type MessageContext struct {
	Session *discordgo.Session
	Message *discordgo.MessageCreate
	Guild   *discordgo.Guild
}

// ReactionContext contains information for reaction events
type ReactionContext struct {
	Session    *discordgo.Session
	Guild      *discordgo.Guild
	Channel    *discordgo.Channel
	Message    *discordgo.Message
	User       *discordgo.User
	Emoji      *discordgo.Emoji
	IsAddition bool
}

// PresenceContext contains information for presence events
type PresenceContext struct {
	Session  *discordgo.Session
	Guild    *discordgo.Guild
	User     *discordgo.User
	Presence *discordgo.PresenceUpdate
}

// ownerID stores the bot owner's user ID
var ownerID string

// SetOwnerID sets the bot owner's user ID (called during startup)
func SetOwnerID(id string) {
	ownerID = id
}
