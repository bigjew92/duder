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
	Execute(ctx *CommandContext) error
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

// CommandContext contains all information for executing a slash command
type CommandContext struct {
	Session     *discordgo.Session
	Interaction *discordgo.InteractionCreate
	Guild       *discordgo.Guild
	Channel     *discordgo.Channel
	User        *discordgo.User
	Options     map[string]*discordgo.ApplicationCommandInteractionDataOption
}

// GetString returns a string option value
func (ctx *CommandContext) GetString(name string) string {
	if opt, ok := ctx.Options[name]; ok {
		return opt.StringValue()
	}
	return ""
}

// GetInt returns an integer option value
func (ctx *CommandContext) GetInt(name string) int64 {
	if opt, ok := ctx.Options[name]; ok {
		return opt.IntValue()
	}
	return 0
}

// GetBool returns a boolean option value
func (ctx *CommandContext) GetBool(name string) bool {
	if opt, ok := ctx.Options[name]; ok {
		return opt.BoolValue()
	}
	return false
}

// GetUser returns a user option value
func (ctx *CommandContext) GetUser(name string) *discordgo.User {
	if opt, ok := ctx.Options[name]; ok {
		return opt.UserValue(ctx.Session)
	}
	return nil
}

// Reply sends a text response to the interaction
func (ctx *CommandContext) Reply(content string) error {
	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
		},
	})
}

// ReplyEmbed sends an embed response to the interaction
func (ctx *CommandContext) ReplyEmbed(embed *discordgo.MessageEmbed) error {
	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Embeds: []*discordgo.MessageEmbed{embed},
		},
	})
}

// ReplyEphemeral sends a private response only visible to the user
func (ctx *CommandContext) ReplyEphemeral(content string) error {
	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseChannelMessageWithSource,
		Data: &discordgo.InteractionResponseData{
			Content: content,
			Flags:   discordgo.MessageFlagsEphemeral,
		},
	})
}

// DeferReply acknowledges the interaction and allows for a delayed response
func (ctx *CommandContext) DeferReply() error {
	return ctx.Session.InteractionRespond(ctx.Interaction.Interaction, &discordgo.InteractionResponse{
		Type: discordgo.InteractionResponseDeferredChannelMessageWithSource,
	})
}

// FollowUp sends a follow-up message after DeferReply
func (ctx *CommandContext) FollowUp(content string) error {
	_, err := ctx.Session.FollowupMessageCreate(ctx.Interaction.Interaction, true, &discordgo.WebhookParams{
		Content: content,
	})
	return err
}

// FollowUpEmbed sends a follow-up embed after DeferReply
func (ctx *CommandContext) FollowUpEmbed(embed *discordgo.MessageEmbed) error {
	_, err := ctx.Session.FollowupMessageCreate(ctx.Interaction.Interaction, true, &discordgo.WebhookParams{
		Embeds: []*discordgo.MessageEmbed{embed},
	})
	return err
}

// IsOwner returns true if the user is the bot owner
func (ctx *CommandContext) IsOwner() bool {
	return ctx.User.ID == ownerID
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
