package rugs

import (
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&StatusCommand{})
	Register(&AvatarCommand{})
}

// StatusCommand implements the /status slash command
type StatusCommand struct{}

func (c *StatusCommand) Name() string {
	return "status"
}

func (c *StatusCommand) Description() string {
	return "Set the bot's status (owner only)"
}

func (c *StatusCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "message",
			Description: "The new status message",
			Required:    true,
		},
	}
}

func (c *StatusCommand) Execute(ctx CommandContext) error {
	if !ctx.IsOwner() {
		return ctx.ReplyEphemeral("You are not authorized to use this command.")
	}

	status := ctx.GetString("message")
	if status == "" {
		return ctx.ReplyEphemeral("Please provide a status message.")
	}

	err := ctx.Session().UpdateGameStatus(0, status)
	if err != nil {
		return ctx.ReplyEphemeral(fmt.Sprintf("Failed to update status: %v", err))
	}

	return ctx.Reply(fmt.Sprintf("Status updated to: %s", status))
}

// AvatarCommand implements the /avatar slash command
type AvatarCommand struct{}

func (c *AvatarCommand) Name() string {
	return "avatar"
}

func (c *AvatarCommand) Description() string {
	return "Manage the bot's avatar (owner only)"
}

func (c *AvatarCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "url",
			Description: "Set avatar from a URL",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "url",
					Description: "URL of the new avatar image",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "save",
			Description: "Save the current avatar",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "filename",
					Description: "Filename to save as (without extension)",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "list",
			Description: "List saved avatars",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "use",
			Description: "Use a saved avatar",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "filename",
					Description: "Filename of the saved avatar",
					Required:    true,
				},
			},
		},
	}
}

func (c *AvatarCommand) Execute(ctx CommandContext) error {
	if !ctx.IsOwner() {
		return ctx.ReplyEphemeral("You are not authorized to use this command.")
	}

	// Determine subcommand from options
	data := ctx.Interaction().ApplicationCommandData()
	if len(data.Options) == 0 {
		return ctx.ReplyEphemeral("Please specify a subcommand.")
	}

	subCmd := data.Options[0].Name

	// Parse subcommand options

	switch subCmd {
	case "url":
		return c.handleURL(ctx)
	case "save":
		return c.handleSave(ctx)
	case "list":
		return c.handleList(ctx)
	case "use":
		return c.handleUse(ctx)
	}

	return ctx.ReplyEphemeral("Unknown subcommand.")
}

func (c *AvatarCommand) handleURL(ctx CommandContext) error {
	url := ctx.GetString("url")
	if url == "" {
		return ctx.ReplyEphemeral("Please provide a URL.")
	}

	// Validate content type
	validTypes := []string{"image/png", "image/jpeg"}
	resp, err := http.Head(url)
	if err != nil {
		return ctx.ReplyEphemeral(fmt.Sprintf("Failed to fetch URL: %v", err))
	}

	contentType := resp.Header.Get("Content-Type")
	valid := false
	for _, t := range validTypes {
		if strings.HasPrefix(contentType, t) {
			valid = true
			break
		}
	}
	if !valid {
		return ctx.ReplyEphemeral(fmt.Sprintf("Invalid content type: %s", contentType))
	}

	// Download the image
	resp, err = http.Get(url)
	if err != nil {
		return ctx.ReplyEphemeral(fmt.Sprintf("Failed to download image: %v", err))
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx.ReplyEphemeral(fmt.Sprintf("Failed to read image: %v", err))
	}

	// Convert to base64
	base64Img := base64.StdEncoding.EncodeToString(data)
	dataURI := fmt.Sprintf("data:%s;base64,%s", contentType, base64Img)

	// Update avatar
	_, err = ctx.Session().UserUpdate("", dataURI, "")
	if err != nil {
		return ctx.ReplyEphemeral(fmt.Sprintf("Failed to update avatar: %v", err))
	}

	return ctx.Reply("Avatar updated successfully!")
}

func (c *AvatarCommand) handleSave(ctx CommandContext) error {
	filename := ctx.GetString("filename")
	if filename == "" {
		return ctx.ReplyEphemeral("Please provide a filename.")
	}

	// Get current avatar URL
	user, err := ctx.Session().User("@me")
	if err != nil {
		return ctx.ReplyEphemeral(fmt.Sprintf("Failed to get bot user: %v", err))
	}

	avatarURL := user.AvatarURL("512")
	resp, err := http.Get(avatarURL)
	if err != nil {
		return ctx.ReplyEphemeral(fmt.Sprintf("Failed to download avatar: %v", err))
	}
	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	if err != nil {
		return ctx.ReplyEphemeral(fmt.Sprintf("Failed to read avatar: %v", err))
	}

	// Save to file
	avatarsPath := "avatars"
	os.MkdirAll(avatarsPath, 0755)

	ext := ".png"
	if strings.Contains(resp.Header.Get("Content-Type"), "jpeg") {
		ext = ".jpg"
	}

	filePath := filepath.Join(avatarsPath, filename+ext)
	if err := os.WriteFile(filePath, data, 0644); err != nil {
		return ctx.ReplyEphemeral(fmt.Sprintf("Failed to save avatar: %v", err))
	}

	return ctx.Reply(fmt.Sprintf("Avatar saved as: %s", filename+ext))
}

func (c *AvatarCommand) handleList(ctx CommandContext) error {
	avatarsPath := "avatars"
	entries, err := os.ReadDir(avatarsPath)
	if err != nil {
		return ctx.ReplyEphemeral("No saved avatars found.")
	}

	var files []string
	for _, entry := range entries {
		if !entry.IsDir() && (strings.HasSuffix(entry.Name(), ".png") || strings.HasSuffix(entry.Name(), ".jpg")) {
			files = append(files, entry.Name())
		}
	}

	if len(files) == 0 {
		return ctx.Reply("No saved avatars found.")
	}

	return ctx.Reply(fmt.Sprintf("Saved avatars:\n```\n%s\n```", strings.Join(files, "\n")))
}

func (c *AvatarCommand) handleUse(ctx CommandContext) error {
	filename := ctx.GetString("filename")
	if filename == "" {
		return ctx.ReplyEphemeral("Please provide a filename.")
	}

	avatarsPath := "avatars"
	filePath := filepath.Join(avatarsPath, filename)

	// Try with extensions if not specified
	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		if _, err := os.Stat(filePath + ".png"); err == nil {
			filePath += ".png"
		} else if _, err := os.Stat(filePath + ".jpg"); err == nil {
			filePath += ".jpg"
		} else {
			return ctx.ReplyEphemeral(fmt.Sprintf("Avatar not found: %s", filename))
		}
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return ctx.ReplyEphemeral(fmt.Sprintf("Failed to read avatar: %v", err))
	}

	contentType := http.DetectContentType(data)
	base64Img := base64.StdEncoding.EncodeToString(data)
	dataURI := fmt.Sprintf("data:%s;base64,%s", contentType, base64Img)

	_, err = ctx.Session().UserUpdate("", dataURI, "")
	if err != nil {
		return ctx.ReplyEphemeral(fmt.Sprintf("Failed to update avatar: %v", err))
	}

	return ctx.Reply(fmt.Sprintf("Now using avatar: %s", filename))
}
