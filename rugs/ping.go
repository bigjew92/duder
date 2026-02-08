package rugs

import (
	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&PingCommand{})
}

// PingCommand implements the /ping slash command
type PingCommand struct{}

// Name returns the command name
func (c *PingCommand) Name() string {
	return "ping"
}

// Description returns the command description
func (c *PingCommand) Description() string {
	return "Returns pong - check if the bot is alive"
}

// Options returns the command options
func (c *PingCommand) Options() []*discordgo.ApplicationCommandOption {
	return nil // No options needed
}

// Execute handles the command
func (c *PingCommand) Execute(ctx *CommandContext) error {
	return ctx.Reply("pong 🏓")
}
