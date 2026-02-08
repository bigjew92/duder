package rugs

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&RaffleCommand{})
}

// RaffleCommand implements the /raffle slash command
type RaffleCommand struct {
	storage *Storage
}

type raffleData struct {
	Active      bool      `json:"active"`
	Description string    `json:"description"`
	Entries     []string  `json:"entries"` // User IDs
	StartedBy   string    `json:"startedBy"`
	StartedAt   time.Time `json:"startedAt"`
}

func (c *RaffleCommand) Name() string {
	return "raffle"
}

func (c *RaffleCommand) Description() string {
	return "Run a raffle in the server"
}

func (c *RaffleCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "start",
			Description: "Start a new raffle",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "description",
					Description: "What's being raffled",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "join",
			Description: "Join the current raffle",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "leave",
			Description: "Leave the current raffle",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "finish",
			Description: "End the raffle and pick a winner",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "cancel",
			Description: "Cancel the current raffle",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "status",
			Description: "Check raffle status",
		},
	}
}

func (c *RaffleCommand) getStorage() *Storage {
	if c.storage == nil {
		c.storage = NewStorage("raffle")
		c.storage.Load()
	}
	return c.storage
}

func (c *RaffleCommand) getRaffle(guildID string) *raffleData {
	storage := c.getStorage()
	guilds := storage.GetMap("guilds")
	if guilds == nil {
		return nil
	}

	data, ok := guilds[guildID].(map[string]interface{})
	if !ok {
		return nil
	}

	raffle := &raffleData{}
	raffle.Active, _ = data["active"].(bool)
	raffle.Description, _ = data["description"].(string)
	raffle.StartedBy, _ = data["startedBy"].(string)

	if entries, ok := data["entries"].([]interface{}); ok {
		for _, e := range entries {
			if s, ok := e.(string); ok {
				raffle.Entries = append(raffle.Entries, s)
			}
		}
	}

	return raffle
}

func (c *RaffleCommand) saveRaffle(guildID string, raffle *raffleData) {
	storage := c.getStorage()

	guilds := storage.GetMap("guilds")
	if guilds == nil {
		guilds = make(map[string]interface{})
	}

	guilds[guildID] = map[string]interface{}{
		"active":      raffle.Active,
		"description": raffle.Description,
		"entries":     raffle.Entries,
		"startedBy":   raffle.StartedBy,
	}

	storage.Set("guilds", guilds)
	storage.Save()
}

func (c *RaffleCommand) Execute(ctx CommandContext) error {
	if ctx.Guild() == nil {
		return ctx.ReplyEphemeral("Raffles can only be run in servers.")
	}

	data := ctx.Interaction().ApplicationCommandData()
	if len(data.Options) == 0 {
		return ctx.ReplyEphemeral("Please specify a subcommand.")
	}

	subCmd := data.Options[0].Name

	switch subCmd {
	case "start":
		return c.handleStart(ctx)
	case "join":
		return c.handleJoin(ctx)
	case "leave":
		return c.handleLeave(ctx)
	case "finish":
		return c.handleFinish(ctx)
	case "cancel":
		return c.handleCancel(ctx)
	case "status":
		return c.handleStatus(ctx)
	}

	return ctx.ReplyEphemeral("Unknown subcommand.")
}

func (c *RaffleCommand) handleStart(ctx CommandContext) error {
	raffle := c.getRaffle(ctx.Guild().ID)
	if raffle != nil && raffle.Active {
		return ctx.ReplyEphemeral("A raffle is already active! Finish it first.")
	}

	description := ctx.GetString("description")
	newRaffle := &raffleData{
		Active:      true,
		Description: description,
		Entries:     []string{},
		StartedBy:   ctx.User().ID,
		StartedAt:   time.Now(),
	}

	c.saveRaffle(ctx.Guild().ID, newRaffle)

	embed := NewEmbed().
		SetTitle("🎉 Raffle Started!").
		SetDescription(description).
		AddField("How to join", "Use `/raffle join` to enter!", false).
		SetColor(ColorGold).
		Build()

	return ctx.ReplyEmbed(embed)
}

func (c *RaffleCommand) handleJoin(ctx CommandContext) error {
	raffle := c.getRaffle(ctx.Guild().ID)
	if raffle == nil || !raffle.Active {
		return ctx.ReplyEphemeral("No active raffle.")
	}

	// Check if already entered
	for _, e := range raffle.Entries {
		if e == ctx.User().ID {
			return ctx.ReplyEphemeral("You're already in the raffle!")
		}
	}

	raffle.Entries = append(raffle.Entries, ctx.User().ID)
	c.saveRaffle(ctx.Guild().ID, raffle)

	return ctx.Reply(fmt.Sprintf("✅ **%s** joined the raffle! (%d entries)", ctx.User().Username, len(raffle.Entries)))
}

func (c *RaffleCommand) handleLeave(ctx CommandContext) error {
	raffle := c.getRaffle(ctx.Guild().ID)
	if raffle == nil || !raffle.Active {
		return ctx.ReplyEphemeral("No active raffle.")
	}

	found := false
	var newEntries []string
	for _, e := range raffle.Entries {
		if e == ctx.User().ID {
			found = true
		} else {
			newEntries = append(newEntries, e)
		}
	}

	if !found {
		return ctx.ReplyEphemeral("You're not in the raffle.")
	}

	raffle.Entries = newEntries
	c.saveRaffle(ctx.Guild().ID, raffle)

	return ctx.Reply(fmt.Sprintf("❌ **%s** left the raffle.", ctx.User().Username))
}

func (c *RaffleCommand) handleFinish(ctx CommandContext) error {
	raffle := c.getRaffle(ctx.Guild().ID)
	if raffle == nil || !raffle.Active {
		return ctx.ReplyEphemeral("No active raffle.")
	}

	if raffle.StartedBy != ctx.User().ID && !ctx.IsOwner() {
		return ctx.ReplyEphemeral("Only the person who started the raffle can finish it.")
	}

	if len(raffle.Entries) == 0 {
		raffle.Active = false
		c.saveRaffle(ctx.Guild().ID, raffle)
		return ctx.Reply("❌ Raffle ended with no entries.")
	}

	// Pick random winner
	winnerIdx := RandomInRange(0, len(raffle.Entries)-1)
	winnerID := raffle.Entries[winnerIdx]

	raffle.Active = false
	c.saveRaffle(ctx.Guild().ID, raffle)

	embed := NewEmbed().
		SetTitle("🎉 Raffle Winner!").
		SetDescription(raffle.Description).
		AddField("Winner", fmt.Sprintf("<@%s>", winnerID), false).
		AddField("Total entries", fmt.Sprintf("%d", len(raffle.Entries)), true).
		SetColor(ColorGold).
		Build()

	return ctx.ReplyEmbed(embed)
}

func (c *RaffleCommand) handleCancel(ctx CommandContext) error {
	raffle := c.getRaffle(ctx.Guild().ID)
	if raffle == nil || !raffle.Active {
		return ctx.ReplyEphemeral("No active raffle.")
	}

	if raffle.StartedBy != ctx.User().ID && !ctx.IsOwner() {
		return ctx.ReplyEphemeral("Only the person who started the raffle can cancel it.")
	}

	raffle.Active = false
	c.saveRaffle(ctx.Guild().ID, raffle)

	return ctx.Reply("🚫 Raffle cancelled.")
}

func (c *RaffleCommand) handleStatus(ctx CommandContext) error {
	raffle := c.getRaffle(ctx.Guild().ID)
	if raffle == nil || !raffle.Active {
		return ctx.Reply("No active raffle.")
	}

	embed := NewEmbed().
		SetTitle("🎟️ Current Raffle").
		SetDescription(raffle.Description).
		AddField("Entries", fmt.Sprintf("%d", len(raffle.Entries)), true).
		AddField("Started by", fmt.Sprintf("<@%s>", raffle.StartedBy), true).
		SetColor(ColorGold).
		Build()

	return ctx.ReplyEmbed(embed)
}
