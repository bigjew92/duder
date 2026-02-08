package rugs

import (
	"fmt"
	"log"
	"sync"

	"github.com/bwmarrin/discordgo"
)

// Registry manages all slash commands
type Registry struct {
	commands         map[string]SlashCommand
	eventHandlers    []EventHandler
	reactionHandlers []ReactionHandler
	presenceHandlers []PresenceHandler
	mu               sync.RWMutex
	session          *discordgo.Session
	guildID          string // Empty for global commands
}

// NewRegistry creates a new command registry
func NewRegistry() *Registry {
	return &Registry{
		commands:         make(map[string]SlashCommand),
		eventHandlers:    make([]EventHandler, 0),
		reactionHandlers: make([]ReactionHandler, 0),
		presenceHandlers: make([]PresenceHandler, 0),
	}
}

// Register adds a command to the registry
func (r *Registry) Register(cmd SlashCommand) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.commands[cmd.Name()] = cmd
	log.Printf("[Registry] Registered command: /%s", cmd.Name())

	// Check if command implements event handlers
	if eh, ok := cmd.(EventHandler); ok {
		r.eventHandlers = append(r.eventHandlers, eh)
		log.Printf("[Registry] Command /%s registered message handler", cmd.Name())
	}
	if rh, ok := cmd.(ReactionHandler); ok {
		r.reactionHandlers = append(r.reactionHandlers, rh)
		log.Printf("[Registry] Command /%s registered reaction handler", cmd.Name())
	}
	if ph, ok := cmd.(PresenceHandler); ok {
		r.presenceHandlers = append(r.presenceHandlers, ph)
		log.Printf("[Registry] Command /%s registered presence handler", cmd.Name())
	}
}

// Get returns a command by name
func (r *Registry) Get(name string) (SlashCommand, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	cmd, ok := r.commands[name]
	return cmd, ok
}

// RegisterWithDiscord registers all commands with the Discord API
func (r *Registry) RegisterWithDiscord(session *discordgo.Session, guildID string) error {
	r.session = session
	r.guildID = guildID

	r.mu.RLock()
	defer r.mu.RUnlock()

	// Build list of application commands
	appCommands := make([]*discordgo.ApplicationCommand, 0, len(r.commands))
	for _, cmd := range r.commands {
		appCmd := &discordgo.ApplicationCommand{
			Name:        cmd.Name(),
			Description: cmd.Description(),
			Options:     cmd.Options(),
		}
		appCommands = append(appCommands, appCmd)
	}

	// Register commands with Discord
	// Use BulkOverwrite to sync all commands at once
	_, err := session.ApplicationCommandBulkOverwrite(session.State.User.ID, guildID, appCommands)
	if err != nil {
		return fmt.Errorf("failed to register commands: %w", err)
	}

	log.Printf("[Registry] Registered %d slash commands with Discord", len(appCommands))
	return nil
}

// HandleInteraction routes an interaction to the appropriate command
func (r *Registry) HandleInteraction(s *discordgo.Session, i *discordgo.InteractionCreate) {
	if i.Type != discordgo.InteractionApplicationCommand {
		return
	}

	data := i.ApplicationCommandData()
	cmd, ok := r.Get(data.Name)
	if !ok {
		log.Printf("[Registry] Unknown command: /%s", data.Name)
		return
	}

	// Build options map
	options := make(map[string]*discordgo.ApplicationCommandInteractionDataOption)
	for _, opt := range data.Options {
		options[opt.Name] = opt
		// Handle subcommand options
		if opt.Type == discordgo.ApplicationCommandOptionSubCommand {
			for _, subOpt := range opt.Options {
				options[subOpt.Name] = subOpt
			}
		}
	}

	// Get guild
	var guild *discordgo.Guild
	if i.GuildID != "" {
		guild, _ = s.Guild(i.GuildID)
	}

	// Get channel
	var channel *discordgo.Channel
	if i.ChannelID != "" {
		channel, _ = s.Channel(i.ChannelID)
	}

	// Get user (from member in guild, or directly in DM)
	var user *discordgo.User
	if i.Member != nil {
		user = i.Member.User
	} else {
		user = i.User
	}

	ctx := &CommandContextImpl{
		SessionVal:     s,
		InteractionVal: i,
		GuildVal:       guild,
		ChannelVal:     channel,
		UserVal:        user,
		OptionsVal:     options,
	}

	if err := cmd.Execute(ctx); err != nil {
		log.Printf("[Registry] Error executing /%s: %v", data.Name, err)
		// Try to respond with error if we haven't responded yet
		ctx.ReplyEphemeral(fmt.Sprintf("Error: %v", err))
	}
}

// HandleMessage dispatches message events to registered handlers
func (r *Registry) HandleMessage(s *discordgo.Session, m *discordgo.MessageCreate, guild *discordgo.Guild) {
	ctx := &MessageContext{
		Session: s,
		Message: m,
		Guild:   guild,
	}

	for _, handler := range r.eventHandlers {
		handler.OnMessage(ctx)
	}
}

// HandleReactionAdd dispatches reaction add events
func (r *Registry) HandleReactionAdd(s *discordgo.Session, guild *discordgo.Guild, channel *discordgo.Channel, message *discordgo.Message, user *discordgo.User, emoji *discordgo.Emoji) {
	ctx := &ReactionContext{
		Session:    s,
		Guild:      guild,
		Channel:    channel,
		Message:    message,
		User:       user,
		Emoji:      emoji,
		IsAddition: true,
	}

	for _, handler := range r.reactionHandlers {
		handler.OnReactionAdd(ctx)
	}
}

// HandleReactionRemove dispatches reaction remove events
func (r *Registry) HandleReactionRemove(s *discordgo.Session, guild *discordgo.Guild, channel *discordgo.Channel, message *discordgo.Message, user *discordgo.User, emoji *discordgo.Emoji) {
	ctx := &ReactionContext{
		Session:    s,
		Guild:      guild,
		Channel:    channel,
		Message:    message,
		User:       user,
		Emoji:      emoji,
		IsAddition: false,
	}

	for _, handler := range r.reactionHandlers {
		handler.OnReactionRemove(ctx)
	}
}

// HandlePresenceUpdate dispatches presence events
func (r *Registry) HandlePresenceUpdate(s *discordgo.Session, guild *discordgo.Guild, user *discordgo.User, presence *discordgo.PresenceUpdate) {
	ctx := &PresenceContext{
		Session:  s,
		Guild:    guild,
		User:     user,
		Presence: presence,
	}

	for _, handler := range r.presenceHandlers {
		handler.OnPresenceUpdate(ctx)
	}
}

// CommandCount returns the number of registered commands
func (r *Registry) CommandCount() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.commands)
}

// DefaultRegistry is the global command registry
var DefaultRegistry = NewRegistry()

// Register adds a command to the default registry
func Register(cmd SlashCommand) {
	DefaultRegistry.Register(cmd)
}
