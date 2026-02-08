package rugs

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&YouTubeCommand{})
}

// YouTubeCommand implements the /youtube slash command
type YouTubeCommand struct {
	storage *Storage
}

func (c *YouTubeCommand) Name() string {
	return "youtube"
}

func (c *YouTubeCommand) Description() string {
	return "Get a random video from a YouTube playlist"
}

func (c *YouTubeCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "random",
			Description: "Get a random video from the configured playlist",
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "setplaylist",
			Description: "Set the playlist ID",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "playlist_id",
					Description: "YouTube playlist ID",
					Required:    true,
				},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionSubCommand,
			Name:        "setkey",
			Description: "Set YouTube API key (owner only)",
			Options: []*discordgo.ApplicationCommandOption{
				{
					Type:        discordgo.ApplicationCommandOptionString,
					Name:        "key",
					Description: "YouTube Data API key",
					Required:    true,
				},
			},
		},
	}
}

func (c *YouTubeCommand) getStorage() *Storage {
	if c.storage == nil {
		c.storage = NewStorage("youtube")
		c.storage.Load()
	}
	return c.storage
}

func (c *YouTubeCommand) getAPIKey() string {
	storage := c.getStorage()
	if key, ok := storage.GetNested("settings", "api_key"); ok {
		if str, ok := key.(string); ok && str != "" {
			return str
		}
	}
	// Fallback to environment variable
	if key := os.Getenv("YOUTUBE_API_KEY"); key != "" {
		return key
	}
	return os.Getenv("GOOGLE_API_KEY")
}

func (c *YouTubeCommand) getPlaylistID() string {
	storage := c.getStorage()
	if id, ok := storage.GetNested("settings", "playlist_id"); ok {
		if str, ok := id.(string); ok {
			return str
		}
	}
	return ""
}

func (c *YouTubeCommand) Execute(ctx CommandContext) error {
	data := ctx.Interaction().ApplicationCommandData()
	if len(data.Options) == 0 {
		return ctx.ReplyEphemeral("Please specify a subcommand.")
	}

	subCmd := data.Options[0].Name

	switch subCmd {
	case "random":
		return c.handleRandom(ctx)
	case "setplaylist":
		return c.handleSetPlaylist(ctx)
	case "setkey":
		return c.handleSetKey(ctx)
	}

	return ctx.ReplyEphemeral("Unknown subcommand.")
}

func (c *YouTubeCommand) handleSetKey(ctx CommandContext) error {
	if !ctx.IsOwner() {
		return ctx.ReplyEphemeral("Only the bot owner can set the API key.")
	}

	key := ctx.GetString("key")
	storage := c.getStorage()
	storage.SetNested(key, "settings", "api_key")
	storage.Save()

	return ctx.ReplyEphemeral("YouTube API key set!")
}

func (c *YouTubeCommand) handleSetPlaylist(ctx CommandContext) error {
	playlistID := ctx.GetString("playlist_id")
	storage := c.getStorage()
	storage.SetNested(playlistID, "settings", "playlist_id")
	storage.Save()

	return ctx.Reply(fmt.Sprintf("Playlist set to: %s", playlistID))
}

func (c *YouTubeCommand) handleRandom(ctx CommandContext) error {
	apiKey := c.getAPIKey()
	if apiKey == "" {
		return ctx.ReplyEphemeral("YouTube API key not configured.")
	}

	playlistID := c.getPlaylistID()
	if playlistID == "" {
		return ctx.ReplyEphemeral("No playlist configured. Use `/youtube setplaylist` first.")
	}

	ctx.DeferReply()

	// Get playlist items
	url := fmt.Sprintf(
		"https://www.googleapis.com/youtube/v3/playlistItems?part=snippet&maxResults=50&playlistId=%s&key=%s",
		playlistID, apiKey)

	resp, err := ctx.HTTPGetString(10, url, nil)
	if err != nil {
		return ctx.FollowUp("Failed to fetch playlist.")
	}

	var result struct {
		Items []struct {
			Snippet struct {
				Title      string `json:"title"`
				ResourceID struct {
					VideoID string `json:"videoId"`
				} `json:"resourceId"`
				Thumbnails struct {
					Default struct {
						URL string `json:"url"`
					} `json:"default"`
				} `json:"thumbnails"`
			} `json:"snippet"`
		} `json:"items"`
	}

	if err := json.Unmarshal([]byte(resp), &result); err != nil || len(result.Items) == 0 {
		return ctx.FollowUp("No videos found in playlist.")
	}

	// Pick random video
	idx := RandomInRange(0, len(result.Items)-1)
	video := result.Items[idx]
	videoURL := fmt.Sprintf("https://www.youtube.com/watch?v=%s", video.Snippet.ResourceID.VideoID)

	embed := NewEmbed().
		SetTitle(fmt.Sprintf("🎬 %s", video.Snippet.Title)).
		SetURL(videoURL).
		SetThumbnail(video.Snippet.Thumbnails.Default.URL).
		SetColor(ColorRed).
		Build()

	return ctx.FollowUpEmbed(embed)
}
