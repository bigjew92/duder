package main

import (
	"log"

	"github.com/bigjew92/duder/rugs"
	"github.com/bwmarrin/discordgo"
)

func init() {
	Duder.Rugs = new(RugManager)
}

// RugManager handles command routing to Go slash commands
type RugManager struct{}

// Load initializes the command system (no-op for Go commands, they self-register)
func (manager *RugManager) Load() error {
	log.Printf("[Rugs] Loaded %d Go slash commands", rugs.DefaultRegistry.CommandCount())
	return nil
}

// OnMessage routes message events to registered handlers
func (manager *RugManager) OnMessage(guild *discordgo.Guild, message *discordgo.MessageCreate) {
	rugs.DefaultRegistry.HandleMessage(Duder.Discord.Session(), message, guild)
}

// OnReactionAdd routes reaction add events to registered handlers
func (manager *RugManager) OnReactionAdd(guild *discordgo.Guild, channel *discordgo.Channel, message *discordgo.Message, user *discordgo.User, emoji *discordgo.Emoji) {
	rugs.DefaultRegistry.HandleReactionAdd(Duder.Discord.Session(), guild, channel, message, user, emoji)
}

// OnReactionRemove routes reaction remove events to registered handlers
func (manager *RugManager) OnReactionRemove(guild *discordgo.Guild, channel *discordgo.Channel, message *discordgo.Message, user *discordgo.User, emoji *discordgo.Emoji) {
	rugs.DefaultRegistry.HandleReactionRemove(Duder.Discord.Session(), guild, channel, message, user, emoji)
}

// OnPresenceUpdate routes presence events to registered handlers
func (manager *RugManager) OnPresenceUpdate(guild *discordgo.Guild, user *discordgo.User, presence *discordgo.PresenceUpdate) {
	rugs.DefaultRegistry.HandlePresenceUpdate(Duder.Discord.Session(), guild, user, presence)
}
