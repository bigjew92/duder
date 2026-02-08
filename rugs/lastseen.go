package rugs

import (
	"fmt"
	"time"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&LastSeenCommand{})
}

// LastSeenCommand implements the /lastseen slash command
type LastSeenCommand struct {
	storage *Storage
}

func (c *LastSeenCommand) Name() string {
	return "lastseen"
}

func (c *LastSeenCommand) Description() string {
	return "Check when a user was last active"
}

func (c *LastSeenCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "The user to check",
			Required:    true,
		},
	}
}

func (c *LastSeenCommand) getStorage() *Storage {
	if c.storage == nil {
		c.storage = NewStorage("lastseen")
		c.storage.Load()
	}
	return c.storage
}

func (c *LastSeenCommand) Execute(ctx CommandContext) error {
	user := ctx.GetUser("user")
	if user == nil {
		return ctx.ReplyEphemeral("Please specify a user.")
	}

	storage := c.getStorage()
	users := storage.GetMap("users")
	if users == nil {
		return ctx.Reply(fmt.Sprintf("No data for %s", user.Username))
	}

	userData, ok := users[user.ID].(map[string]interface{})
	if !ok {
		return ctx.Reply(fmt.Sprintf("No data for %s", user.Username))
	}

	lastSeenMs, ok := userData["lastSeen"].(float64)
	if !ok {
		return ctx.Reply(fmt.Sprintf("No data for %s", user.Username))
	}

	lastSeen := time.UnixMilli(int64(lastSeenMs))
	duration := time.Since(lastSeen)

	var timeStr string
	switch {
	case duration < time.Minute:
		timeStr = "just now"
	case duration < time.Hour:
		timeStr = fmt.Sprintf("%d minutes ago", int(duration.Minutes()))
	case duration < 24*time.Hour:
		timeStr = fmt.Sprintf("%d hours ago", int(duration.Hours()))
	default:
		timeStr = fmt.Sprintf("%d days ago", int(duration.Hours()/24))
	}

	return ctx.Reply(fmt.Sprintf("**%s** was last seen %s", user.Username, timeStr))
}

// OnMessage implements EventHandler to track user activity
func (c *LastSeenCommand) OnMessage(ctx *MessageContext) {
	if ctx.Message.Author.Bot {
		return
	}

	storage := c.getStorage()

	users := storage.GetMap("users")
	if users == nil {
		users = make(map[string]interface{})
	}

	users[ctx.Message.Author.ID] = map[string]interface{}{
		"lastSeen": time.Now().UnixMilli(),
	}

	storage.Set("users", users)
	storage.Save()
}
