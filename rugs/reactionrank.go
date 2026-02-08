package rugs

import (
	"fmt"
	"sort"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&RankCommand{})
}

// RankCommand implements the /rank slash command with XP tracking
type RankCommand struct {
	storage *Storage
}

func (c *RankCommand) Name() string {
	return "rank"
}

func (c *RankCommand) Description() string {
	return "Check your reaction XP rank"
}

func (c *RankCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionUser,
			Name:        "user",
			Description: "User to check (optional)",
			Required:    false,
		},
	}
}

func (c *RankCommand) getStorage() *Storage {
	if c.storage == nil {
		c.storage = NewStorage("reactionrank")
		c.storage.Load()
	}
	return c.storage
}

func (c *RankCommand) getUserXP(guildID, userID string) int {
	storage := c.getStorage()
	if xp, ok := storage.GetNested("guilds", guildID, userID, "xp"); ok {
		if f, ok := xp.(float64); ok {
			return int(f)
		}
	}
	return 0
}

func (c *RankCommand) setUserXP(guildID, userID string, xp int) {
	storage := c.getStorage()
	storage.SetNested(xp, "guilds", guildID, userID, "xp")
	storage.Save()
}

func (c *RankCommand) Execute(ctx CommandContext) error {
	if ctx.Guild() == nil {
		return ctx.ReplyEphemeral("Ranks are server-specific.")
	}

	targetUser := ctx.GetUser("user")
	if targetUser == nil {
		targetUser = ctx.User()
	}

	xp := c.getUserXP(ctx.Guild().ID, targetUser.ID)
	level := xp / 100

	embed := NewEmbed().
		SetTitle(fmt.Sprintf("📊 %s's Rank", targetUser.Username)).
		AddField("XP", fmt.Sprintf("%d", xp), true).
		AddField("Level", fmt.Sprintf("%d", level), true).
		AddField("Next Level", fmt.Sprintf("%d/%d XP", xp%100, 100), true).
		SetColor(ColorTeal).
		SetThumbnail(targetUser.AvatarURL("128")).
		Build()

	return ctx.ReplyEmbed(embed)
}

// OnReactionAdd implements ReactionHandler
func (c *RankCommand) OnReactionAdd(ctx *ReactionContext) {
	if ctx.User.Bot || ctx.Guild == nil {
		return
	}

	// Get the message author's ID and give them XP
	if ctx.Message != nil && ctx.Message.Author != nil && ctx.Message.Author.ID != ctx.User.ID {
		authorID := ctx.Message.Author.ID
		xp := c.getUserXP(ctx.Guild.ID, authorID)
		c.setUserXP(ctx.Guild.ID, authorID, xp+10)
	}
}

// OnReactionRemove implements ReactionHandler
func (c *RankCommand) OnReactionRemove(ctx *ReactionContext) {
	if ctx.User.Bot || ctx.Guild == nil {
		return
	}

	// Remove XP when reaction is removed
	if ctx.Message != nil && ctx.Message.Author != nil && ctx.Message.Author.ID != ctx.User.ID {
		authorID := ctx.Message.Author.ID
		xp := c.getUserXP(ctx.Guild.ID, authorID)
		if xp >= 10 {
			c.setUserXP(ctx.Guild.ID, authorID, xp-10)
		}
	}
}

// LeaderboardCommand implements the /leaderboard slash command
type LeaderboardCommand struct {
	rankCmd *RankCommand
}

func init() {
	Register(&LeaderboardCommand{})
}

func (c *LeaderboardCommand) Name() string {
	return "leaderboard"
}

func (c *LeaderboardCommand) Description() string {
	return "Show the server's XP leaderboard"
}

func (c *LeaderboardCommand) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

func (c *LeaderboardCommand) getRankStorage() *Storage {
	if c.rankCmd == nil {
		c.rankCmd = &RankCommand{}
	}
	return c.rankCmd.getStorage()
}

func (c *LeaderboardCommand) Execute(ctx CommandContext) error {
	if ctx.Guild() == nil {
		return ctx.ReplyEphemeral("Leaderboards are server-specific.")
	}

	storage := c.getRankStorage()
	guildData, ok := storage.GetNested("guilds", ctx.Guild().ID)
	if !ok {
		return ctx.Reply("No rankings yet!")
	}

	users, ok := guildData.(map[string]interface{})
	if !ok {
		return ctx.Reply("No rankings yet!")
	}

	// Build sorted leaderboard
	type entry struct {
		userID string
		xp     int
	}

	var entries []entry
	for userID, data := range users {
		if userData, ok := data.(map[string]interface{}); ok {
			if xp, ok := userData["xp"].(float64); ok {
				entries = append(entries, entry{userID, int(xp)})
			}
		}
	}

	sort.Slice(entries, func(i, j int) bool {
		return entries[i].xp > entries[j].xp
	})

	// Build embed
	embed := NewEmbed().
		SetTitle(fmt.Sprintf("🏆 %s Leaderboard", ctx.Guild().Name)).
		SetColor(ColorGold)

	for i := 0; i < 10 && i < len(entries); i++ {
		medal := ""
		switch i {
		case 0:
			medal = "🥇"
		case 1:
			medal = "🥈"
		case 2:
			medal = "🥉"
		default:
			medal = fmt.Sprintf("%d.", i+1)
		}

		embed.AddField(
			fmt.Sprintf("%s <@%s>", medal, entries[i].userID),
			fmt.Sprintf("%d XP (Level %d)", entries[i].xp, entries[i].xp/100),
			false,
		)
	}

	return ctx.ReplyEmbed(embed.Build())
}
