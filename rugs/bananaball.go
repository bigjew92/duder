package rugs

import (
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&BananaBallCommand{})
}

// Team colors and emojis
var teamColors = map[string]int{
	"savannah bananas":      0xFFD700, // Yellow
	"party animals":         0xFF69B4, // Pink
	"firefighters":          0xFF0000, // Red
	"texas tailgaters":      0x162751, // Dark Blue
	"banana ball all-stars": 0xDAA520, // Gold
}

var teamEmojis = map[string]string{
	"savannah bananas":      "🍌",
	"party animals":         "🐵",
	"firefighters":          "👨‍🚒",
	"texas tailgaters":      "🍴",
	"banana ball all-stars": "✨",
}

// BananaBallCommand implements the /bananaball slash command
type BananaBallCommand struct{}

func (c *BananaBallCommand) Name() string {
	return "bananaball"
}

func (c *BananaBallCommand) Description() string {
	return "Lookup Banana Ball player or team stats"
}

func (c *BananaBallCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "name",
			Description: "Player or team name",
			Required:    true,
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "type",
			Description: "Search type (default: player)",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "Player", Value: "player"},
				{Name: "Team", Value: "team"},
			},
		},
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "stats",
			Description: "Stats category for players (default: hitting)",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "Hitting", Value: "hitting"},
				{Name: "Pitching", Value: "pitching"},
			},
		},
	}
}

func (c *BananaBallCommand) Execute(ctx CommandContext) error {
	searchType := ctx.GetString("type")
	if searchType == "" {
		searchType = "player"
	}
	name := ctx.GetString("name")
	ctx.DeferReply()

	if searchType == "team" {
		return c.handleTeam(ctx, name)
	}

	statsCategory := ctx.GetString("stats")
	if statsCategory == "" {
		statsCategory = "hitting"
	}
	return c.handlePlayer(ctx, name, statsCategory)
}

func (c *BananaBallCommand) handlePlayer(ctx CommandContext, playerName string, category string) error {
	apiURL := fmt.Sprintf("https://stats.bananaball.com/stats/players_stats?category=%s", url.QueryEscape(category))
	resp, err := ctx.HTTPGetString(10, apiURL, nil)
	if err != nil {
		return ctx.FollowUp("Failed to fetch Banana Ball stats.")
	}

	type PlayerStats struct {
		FirstName       string `json:"first_name"`
		LastName        string `json:"last_name"`
		JerseyNumber    int    `json:"jersey_number"`
		PrimaryPosition struct {
			Label string `json:"label"`
		} `json:"primary_position"`
		Image           string  `json:"image"`
		GamesPlayed     int     `json:"games_played"`
		AtBats          int     `json:"at_bats"`
		BattingAverage  float64 `json:"batting_average"`
		HomeRuns        int     `json:"home_runs"`
		RBI             int     `json:"runs_batted_in"`
		OPS             float64 `json:"on_base_plus_slugging"`
		BallFourSprints int     `json:"ball_four_sprints"`
		// Pitching stats
		InningsPitched float64 `json:"innings_pitched"`
		ERA            float64 `json:"earned_run_average"`
		Wins           int     `json:"wins"`
		Losses         int     `json:"losses"`
		Strikeouts     int     `json:"pitcher_strikeouts"`
		SHIP           float64 `json:"ship"`
	}

	var players []PlayerStats
	if err := json.Unmarshal([]byte(resp), &players); err != nil {
		return ctx.FollowUp("Failed to parse player stats.")
	}

	// Find player by name (fuzzy match)
	nameLower := strings.ToLower(playerName)
	var match *PlayerStats

	for _, p := range players {
		fullName := strings.ToLower(p.FirstName + " " + p.LastName)
		if strings.Contains(fullName, nameLower) ||
			strings.Contains(strings.ToLower(p.LastName), nameLower) ||
			strings.Contains(strings.ToLower(p.FirstName), nameLower) {
			match = &p
			break
		}
	}

	if match == nil {
		return ctx.FollowUp(fmt.Sprintf("No Banana Ball player found matching '%s'.", playerName))
	}

	name := fmt.Sprintf("%s %s", match.FirstName, match.LastName)
	embed := NewEmbed().
		SetTitle(fmt.Sprintf("🍌 %s (#%d)", name, match.JerseyNumber)).
		SetDescription(match.PrimaryPosition.Label).
		SetColor(0xFFD700)

	if match.Image != "" {
		embed.SetThumbnail(fmt.Sprintf("https://banana-stats-pages.vercel.app/_next/image?url=https://res.cloudinary.com/dkrtbxpwe/image/upload/v1/%s&w=256&q=75", match.Image))
	}

	if category == "pitching" {
		embed.AddField("G", fmt.Sprintf("%d", match.GamesPlayed), true)
		embed.AddField("IP", fmt.Sprintf("%.1f", match.InningsPitched), true)
		embed.AddField("ERA", fmt.Sprintf("%.2f", match.ERA), true)
		embed.AddField("W-L", fmt.Sprintf("%d-%d", match.Wins, match.Losses), true)
		embed.AddField("SO", fmt.Sprintf("%d", match.Strikeouts), true)
		embed.AddField("SHIP", fmt.Sprintf("%.2f", match.SHIP), true)
	} else {
		embed.AddField("G", fmt.Sprintf("%d", match.GamesPlayed), true)
		embed.AddField("AB", fmt.Sprintf("%d", match.AtBats), true)
		embed.AddField("AVG", fmt.Sprintf("%.3f", match.BattingAverage), true)
		embed.AddField("HR", fmt.Sprintf("%d", match.HomeRuns), true)
		embed.AddField("RBI", fmt.Sprintf("%d", match.RBI), true)
		embed.AddField("OPS", fmt.Sprintf("%.3f", match.OPS), true)
		embed.AddField("B4S", fmt.Sprintf("%d", match.BallFourSprints), true)
	}

	return ctx.FollowUpEmbed(embed.Build())
}

func (c *BananaBallCommand) handleTeam(ctx CommandContext, teamName string) error {
	apiURL := "https://stats.bananaball.com/stats/teams_stats?category=teams"
	resp, err := ctx.HTTPGetString(10, apiURL, nil)
	if err != nil {
		return ctx.FollowUp("Failed to fetch Banana Ball team stats.")
	}

	type TeamStats struct {
		ID           string `json:"id"`
		Name         string `json:"name"`
		Abbreviation string `json:"abbreviation"`
		Logo         string `json:"logo"`
		Record       struct {
			Wins   int `json:"wins"`
			Losses int `json:"losses"`
		} `json:"record"`
		BattingAverage  float64 `json:"batting_average"`
		ERA             float64 `json:"earned_run_average"`
		HomeRuns        int     `json:"home_runs"`
		BallFourSprints int     `json:"ball_four_sprints"`
	}

	var teams []TeamStats
	if err := json.Unmarshal([]byte(resp), &teams); err != nil {
		return ctx.FollowUp("Failed to parse team stats.")
	}

	// Find team by name (fuzzy match)
	nameLower := strings.ToLower(teamName)
	var match *TeamStats

	for _, t := range teams {
		if strings.Contains(strings.ToLower(t.Name), nameLower) ||
			strings.EqualFold(t.Abbreviation, teamName) {
			match = &t
			break
		}
	}

	if match == nil {
		return ctx.FollowUp(fmt.Sprintf("No Banana Ball team found matching '%s'.", teamName))
	}

	// Get team color and emoji
	teamNameLower := strings.ToLower(match.Name)
	color := 0xFFD700 // Default yellow
	if c, ok := teamColors[teamNameLower]; ok {
		color = c
	}
	emoji := "🍌" // Default
	if e, ok := teamEmojis[teamNameLower]; ok {
		emoji = e
	}

	// Filter to only main teams
	mainTeams := map[string]bool{
		"savannah bananas": true,
		"party animals":    true,
		"firefighters":     true,
		"texas tailgaters": true,
	}

	var filteredTeams []TeamStats
	for _, t := range teams {
		if mainTeams[strings.ToLower(t.Name)] {
			filteredTeams = append(filteredTeams, t)
		}
	}

	// Build league standings table
	sort.Slice(filteredTeams, func(i, j int) bool {
		iWinPct := float64(filteredTeams[i].Record.Wins) / float64(filteredTeams[i].Record.Wins+filteredTeams[i].Record.Losses)
		jWinPct := float64(filteredTeams[j].Record.Wins) / float64(filteredTeams[j].Record.Wins+filteredTeams[j].Record.Losses)
		return iWinPct > jWinPct
	})

	var standingsTable strings.Builder
	standingsTable.WriteString("```\n")
	standingsTable.WriteString(fmt.Sprintf("%-24s %7s\n", "Team", "W-L"))
	standingsTable.WriteString(strings.Repeat("-", 33) + "\n")
	for _, t := range filteredTeams {
		record := fmt.Sprintf("%d-%d", t.Record.Wins, t.Record.Losses)
		marker := " "
		teamEmoji := ""
		if e, ok := teamEmojis[strings.ToLower(t.Name)]; ok {
			teamEmoji = e
		}
		if t.ID == match.ID {
			marker = "▶"
		}
		standingsTable.WriteString(fmt.Sprintf("%s%s %-21s %7s\n", marker, teamEmoji, t.Name, record))
	}
	standingsTable.WriteString("```")

	embed := NewEmbed().
		SetTitle(fmt.Sprintf("%s %s", emoji, match.Name)).
		SetDescription("2025 World Tour").
		SetColor(color)

	if match.Logo != "" {
		embed.SetThumbnail(fmt.Sprintf("https://banana-stats-pages.vercel.app/_next/image?url=https://res.cloudinary.com/dkrtbxpwe/image/upload/v1/%s&w=256&q=75", match.Logo))
	}

	embed.AddField("Record", fmt.Sprintf("%d-%d", match.Record.Wins, match.Record.Losses), true)
	embed.AddField("AVG", fmt.Sprintf("%.3f", match.BattingAverage), true)
	embed.AddField("ERA", fmt.Sprintf("%.2f", match.ERA), true)
	embed.AddField("HR", fmt.Sprintf("%d", match.HomeRuns), true)
	embed.AddField("B4S", fmt.Sprintf("%d", match.BallFourSprints), true)
	embed.AddField("League Standings", standingsTable.String(), false)

	return ctx.FollowUpEmbed(embed.Build())
}
