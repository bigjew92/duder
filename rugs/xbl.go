package rugs

import (
	"encoding/json"
	"fmt"
	"net/url"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&XBLCommand{})
}

// XBLCommand implements the /xbl slash command
type XBLCommand struct {
	storage *Storage
}

func (c *XBLCommand) Name() string {
	return "xbl"
}

func (c *XBLCommand) Description() string {
	return "Xbox Live gamertag lookup"
}

func (c *XBLCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "profile",
			Description: "Look up an Xbox Live profile",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "gamertag",
					Description: "Xbox gamertag (or leave empty for saved gamertag)",
					Required:    false,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "setgamertag",
			Description: "Set your default gamertag",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "gamertag",
					Description: "Your Xbox gamertag",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "setkey",
			Description: "Set OpenXBL API key (owner only)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "key",
					Description: "OpenXBL API key",
					Required:    true,
				},
			},
		},
	}
}

func (c *XBLCommand) getStorage() *Storage {
	if c.storage == nil {
		c.storage = NewStorage("xbl")
		c.storage.Load()
	}
	return c.storage
}

func (c *XBLCommand) getAPIKey() string {
	storage := c.getStorage()
	if key, ok := storage.GetNested("settings", "api_key"); ok {
		if str, ok := key.(string); ok {
			return str
		}
	}
	return ""
}

func (c *XBLCommand) Execute(ctx *CommandContext) error {
	data := ctx.Interaction.ApplicationCommandData()
	if len(data.Options) == 0 {
		return ctx.ReplyEphemeral("Please specify a subcommand.")
	}

	subCmd := data.Options[0].Name
	for _, opt := range data.Options[0].Options {
		ctx.Options[opt.Name] = opt
	}

	switch subCmd {
	case "profile":
		return c.handleProfile(ctx)
	case "setgamertag":
		return c.handleSetGamertag(ctx)
	case "setkey":
		return c.handleSetKey(ctx)
	}

	return ctx.ReplyEphemeral("Unknown subcommand.")
}

func (c *XBLCommand) handleSetKey(ctx *CommandContext) error {
	if !ctx.IsOwner() {
		return ctx.ReplyEphemeral("Only the bot owner can set the API key.")
	}

	key := ctx.GetString("key")
	storage := c.getStorage()
	storage.SetNested(key, "settings", "api_key")
	storage.Save()

	return ctx.ReplyEphemeral("OpenXBL API key set!")
}

func (c *XBLCommand) handleSetGamertag(ctx *CommandContext) error {
	gamertag := ctx.GetString("gamertag")
	storage := c.getStorage()
	storage.SetNested(gamertag, "users", ctx.User.ID, "gamertag")
	storage.Save()

	return ctx.Reply(fmt.Sprintf("Gamertag set to: %s", gamertag))
}

func (c *XBLCommand) handleProfile(ctx *CommandContext) error {
	apiKey := c.getAPIKey()
	if apiKey == "" {
		return ctx.ReplyEphemeral("Xbox API key not configured.")
	}

	gamertag := ctx.GetString("gamertag")
	if gamertag == "" {
		storage := c.getStorage()
		if saved, ok := storage.GetNested("users", ctx.User.ID, "gamertag"); ok {
			if str, ok := saved.(string); ok {
				gamertag = str
			}
		}
	}

	if gamertag == "" {
		return ctx.ReplyEphemeral("Please provide a gamertag or set a default with `/xbl setgamertag`.")
	}

	ctx.DeferReply()

	// Look up XUID first
	xuidURL := fmt.Sprintf("https://xbl.io/api/v2/friends/search?gt=%s", url.QueryEscape(gamertag))
	headers := map[string]string{
		"X-Authorization": apiKey,
	}

	resp, err := HTTPGetString(10, xuidURL, headers)
	if err != nil {
		return ctx.FollowUp("Failed to look up gamertag.")
	}

	var xuidResult struct {
		ProfileUsers []struct {
			ID       string `json:"id"`
			Settings []struct {
				ID    string `json:"id"`
				Value string `json:"value"`
			} `json:"settings"`
		} `json:"profileUsers"`
	}

	if err := json.Unmarshal([]byte(resp), &xuidResult); err != nil || len(xuidResult.ProfileUsers) == 0 {
		return ctx.FollowUp("Gamertag not found.")
	}

	user := xuidResult.ProfileUsers[0]

	// Build embed from settings
	embed := NewEmbed().
		SetTitle(fmt.Sprintf("🎮 Xbox Profile")).
		SetColor(ColorGreen)

	var gamerPic, gamerScore, accountTier, gamertagDisplay string
	for _, setting := range user.Settings {
		switch setting.ID {
		case "Gamertag":
			gamertagDisplay = setting.Value
		case "GameDisplayPicRaw":
			gamerPic = setting.Value
		case "Gamerscore":
			gamerScore = setting.Value
		case "AccountTier":
			accountTier = setting.Value
		}
	}

	if gamertagDisplay != "" {
		embed.SetTitle(fmt.Sprintf("🎮 %s", gamertagDisplay))
	}
	if gamerPic != "" {
		embed.SetThumbnail(gamerPic)
	}
	if gamerScore != "" {
		embed.AddField("Gamerscore", gamerScore, true)
	}
	if accountTier != "" {
		embed.AddField("Account Tier", accountTier, true)
	}
	embed.AddField("XUID", user.ID, false)

	return ctx.FollowUpEmbed(embed.Build())
}
