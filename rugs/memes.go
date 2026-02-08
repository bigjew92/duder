package rugs

import (
	"encoding/json"
	"fmt"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&MemeCommand{})
}

// MemeCommand implements the /meme slash command
type MemeCommand struct {
	storage *Storage
	memes   []memeTemplate
}

type memeTemplate struct {
	ID   string `json:"id"`
	Name string `json:"name"`
}

func (c *MemeCommand) Name() string {
	return "meme"
}

func (c *MemeCommand) Description() string {
	return "Create a meme using Imgflip"
}

func (c *MemeCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "create",
			Description: "Create a meme",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "template",
					Description: "Template name (use /meme list to see options)",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "top",
					Description: "Top text",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "bottom",
					Description: "Bottom text",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "list",
			Description: "List available meme templates",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "setcreds",
			Description: "Set Imgflip credentials (owner only)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "username",
					Description: "Imgflip username",
					Required:    true,
				},
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "password",
					Description: "Imgflip password",
					Required:    true,
				},
			},
		},
	}
}

func (c *MemeCommand) getStorage() *Storage {
	if c.storage == nil {
		c.storage = NewStorage("memes")
		c.storage.Load()
	}
	return c.storage
}

func (c *MemeCommand) loadTemplates() {
	if len(c.memes) > 0 {
		return
	}

	resp, err := HTTPGetString(10, "https://api.imgflip.com/get_memes", nil)
	if err != nil {
		return
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			Memes []memeTemplate `json:"memes"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return
	}

	c.memes = result.Data.Memes
}

func (c *MemeCommand) Execute(ctx *CommandContext) error {
	data := ctx.Interaction.ApplicationCommandData()
	if len(data.Options) == 0 {
		return ctx.ReplyEphemeral("Please specify a subcommand.")
	}

	subCmd := data.Options[0].Name
	for _, opt := range data.Options[0].Options {
		ctx.Options[opt.Name] = opt
	}

	switch subCmd {
	case "create":
		return c.handleCreate(ctx)
	case "list":
		return c.handleList(ctx)
	case "setcreds":
		return c.handleSetCreds(ctx)
	}

	return ctx.ReplyEphemeral("Unknown subcommand.")
}

func (c *MemeCommand) handleSetCreds(ctx *CommandContext) error {
	if !ctx.IsOwner() {
		return ctx.ReplyEphemeral("Only the bot owner can set credentials.")
	}

	username := ctx.GetString("username")
	password := ctx.GetString("password")

	storage := c.getStorage()
	storage.SetNested(username, "settings", "username")
	storage.SetNested(password, "settings", "password")
	storage.Save()

	return ctx.ReplyEphemeral("Imgflip credentials saved!")
}

func (c *MemeCommand) handleList(ctx *CommandContext) error {
	c.loadTemplates()

	if len(c.memes) == 0 {
		return ctx.ReplyEphemeral("Failed to load meme templates.")
	}

	// Show first 20 templates
	var list string
	count := 20
	if len(c.memes) < count {
		count = len(c.memes)
	}
	for i := 0; i < count; i++ {
		list += fmt.Sprintf("• %s\n", c.memes[i].Name)
	}

	return ctx.Reply(fmt.Sprintf("**Popular meme templates:**\n%s", list))
}

func (c *MemeCommand) handleCreate(ctx *CommandContext) error {
	storage := c.getStorage()

	username, _ := storage.GetNested("settings", "username")
	password, _ := storage.GetNested("settings", "password")

	if username == nil || password == nil {
		return ctx.ReplyEphemeral("Imgflip credentials not configured.")
	}

	templateName := ctx.GetString("template")
	topText := ctx.GetString("top")
	bottomText := ctx.GetString("bottom")

	c.loadTemplates()

	// Find template ID
	var templateID string
	for _, m := range c.memes {
		if m.Name == templateName {
			templateID = m.ID
			break
		}
	}

	if templateID == "" {
		return ctx.ReplyEphemeral(fmt.Sprintf("Template '%s' not found. Use `/meme list` to see options.", templateName))
	}

	ctx.DeferReply()

	// Create meme
	resp, err := HTTPPostString(10, "https://api.imgflip.com/caption_image", map[string]string{
		"template_id": templateID,
		"username":    username.(string),
		"password":    password.(string),
		"text0":       topText,
		"text1":       bottomText,
	})
	if err != nil {
		return ctx.FollowUp("Failed to create meme.")
	}

	var result struct {
		Success bool `json:"success"`
		Data    struct {
			URL string `json:"url"`
		} `json:"data"`
		ErrorMessage string `json:"error_message"`
	}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return ctx.FollowUp("Failed to parse response.")
	}

	if !result.Success {
		return ctx.FollowUp(fmt.Sprintf("Imgflip error: %s", result.ErrorMessage))
	}

	embed := NewEmbed().
		SetTitle(templateName).
		SetImage(result.Data.URL).
		SetColor(ColorPurple).
		Build()

	return ctx.FollowUpEmbed(embed)
}
