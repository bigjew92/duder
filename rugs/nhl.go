package rugs

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&NHLCommand{})
}

// NHLCommand implements the /nhl slash command
type NHLCommand struct{}

func (c *NHLCommand) Name() string {
	return "nhl"
}

func (c *NHLCommand) Description() string {
	return "Lookup NHL player or team stats"
}

func (c *NHLCommand) Options() []*discordgo.ApplicationCommandOption {
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
			Description: "Search type (default: team)",
			Required:    false,
			Choices: []*discordgo.ApplicationCommandOptionChoice{
				{Name: "Team", Value: "team"},
				{Name: "Player", Value: "player"},
			},
		},
	}
}

func (c *NHLCommand) Execute(ctx CommandContext) error {
	searchType := ctx.GetString("type")
	if searchType == "" {
		searchType = "team"
	}
	name := ctx.GetString("name")
	ctx.DeferReply()

	if searchType == "team" {
		return c.handleTeam(ctx, name)
	}
	return c.handlePlayer(ctx, name)
}

func (c *NHLCommand) handlePlayer(ctx CommandContext, playerName string) error {
	searchURL := fmt.Sprintf("https://search.d3.nhle.com/api/v1/search/player?culture=en-us&limit=5&q=%s", url.QueryEscape(playerName))
	searchResp, err := ctx.HTTPGetString(10, searchURL, nil)
	if err != nil {
		return ctx.FollowUp("Failed to contact NHL API.")
	}

	type SearchResult struct {
		PlayerID  int `json:"playerId"`
		FirstName struct {
			Default string `json:"default"`
		} `json:"firstName"`
		LastName struct {
			Default string `json:"default"`
		} `json:"lastName"`
		TeamAbbrev string `json:"teamAbbrev"`
		Position   string `json:"position"`
	}

	var results []SearchResult
	if err := json.Unmarshal([]byte(searchResp), &results); err != nil {
		return ctx.FollowUp("Failed to parse search results.")
	}

	if len(results) == 0 {
		return ctx.FollowUp("No NHL players found.")
	}

	player := results[0]

	landingURL := fmt.Sprintf("https://api-web.nhle.com/v1/player/%d/landing", player.PlayerID)
	landingResp, err := ctx.HTTPGetString(10, landingURL, nil)
	if err != nil {
		return ctx.FollowUp("Failed to fetch player stats.")
	}

	type FeaturedStats struct {
		Season      int     `json:"season"`
		Goals       int     `json:"goals"`
		Assists     int     `json:"assists"`
		Points      int     `json:"points"`
		PlusMinus   int     `json:"plusMinus"`
		GamesPlayed int     `json:"gamesPlayed"`
		Wins        int     `json:"wins"`
		Losses      int     `json:"losses"`
		OTLosses    int     `json:"otLosses"`
		SavePctg    float64 `json:"savePctg"`
		GAA         float64 `json:"goalsAgainstAvg"`
	}

	var landing struct {
		FeaturedStats struct {
			RegularSeason struct {
				SubSeason FeaturedStats `json:"subSeason"`
			} `json:"regularSeason"`
		} `json:"featuredStats"`
		Position string `json:"position"`
		Headshot string `json:"headshot"`
	}

	if err := json.Unmarshal([]byte(landingResp), &landing); err != nil {
		return ctx.FollowUp("Failed to parse player details.")
	}

	stats := landing.FeaturedStats.RegularSeason.SubSeason
	name := fmt.Sprintf("%s %s", player.FirstName.Default, player.LastName.Default)
	embed := NewEmbed().
		SetTitle(fmt.Sprintf("%s (%s)", name, player.Position)).
		SetDescription(fmt.Sprintf("Season %d-%d", stats.Season/10000, stats.Season%10000)).
		SetColor(ColorBlue).
		SetThumbnail(landing.Headshot)

	if landing.Position == "G" {
		embed.AddField("GP", fmt.Sprintf("%d", stats.GamesPlayed), true)
		embed.AddField("Record", fmt.Sprintf("%d-%d-%d", stats.Wins, stats.Losses, stats.OTLosses), true)
		embed.AddField("GAA", fmt.Sprintf("%.2f", stats.GAA), true)
		embed.AddField("SV%", fmt.Sprintf("%.3f", stats.SavePctg), true)
	} else {
		embed.AddField("Goals", fmt.Sprintf("%d", stats.Goals), true)
		embed.AddField("Assists", fmt.Sprintf("%d", stats.Assists), true)
		embed.AddField("Points", fmt.Sprintf("%d", stats.Points), true)
		if stats.PlusMinus > 0 {
			embed.AddField("+/-", fmt.Sprintf("+%d", stats.PlusMinus), true)
		} else {
			embed.AddField("+/-", fmt.Sprintf("%d", stats.PlusMinus), true)
		}
	}

	return ctx.FollowUpEmbed(embed.Build())
}

func (c *NHLCommand) handleTeam(ctx CommandContext, teamName string) error {
	// Get current standings
	today := time.Now().Format("2006-01-02")
	standingsURL := fmt.Sprintf("https://api-web.nhle.com/v1/standings/%s", today)
	resp, err := ctx.HTTPGetString(10, standingsURL, nil)
	if err != nil {
		return ctx.FollowUp("Failed to fetch NHL standings.")
	}

	type TeamStanding struct {
		TeamName struct {
			Default string `json:"default"`
		} `json:"teamName"`
		TeamAbbrev struct {
			Default string `json:"default"`
		} `json:"teamAbbrev"`
		TeamLogo         string  `json:"teamLogo"`
		ConferenceName   string  `json:"conferenceName"`
		DivisionName     string  `json:"divisionName"`
		GamesPlayed      int     `json:"gamesPlayed"`
		Wins             int     `json:"wins"`
		Losses           int     `json:"losses"`
		OTLosses         int     `json:"otLosses"`
		Points           int     `json:"points"`
		PointPctg        float64 `json:"pointPctg"`
		GoalFor          int     `json:"goalFor"`
		GoalAgainst      int     `json:"goalAgainst"`
		GoalDifferential int     `json:"goalDifferential"`
		StreakCode       string  `json:"streakCode"`
		StreakCount      int     `json:"streakCount"`
	}

	var standings struct {
		Standings []TeamStanding `json:"standings"`
	}

	if err := json.Unmarshal([]byte(resp), &standings); err != nil {
		return ctx.FollowUp("Failed to parse standings.")
	}

	// Find team by name (fuzzy match)
	nameLower := strings.ToLower(teamName)
	var match *TeamStanding

	for _, team := range standings.Standings {
		if strings.Contains(strings.ToLower(team.TeamName.Default), nameLower) ||
			strings.EqualFold(team.TeamAbbrev.Default, teamName) {
			match = &team
			break
		}
	}

	if match == nil {
		return ctx.FollowUp(fmt.Sprintf("No NHL team found matching '%s'.", teamName))
	}

	// Build division standings table
	var divisionTeams []TeamStanding
	for _, team := range standings.Standings {
		if team.DivisionName == match.DivisionName {
			divisionTeams = append(divisionTeams, team)
		}
	}

	var standingsTable strings.Builder
	standingsTable.WriteString("```\n")
	standingsTable.WriteString(fmt.Sprintf("%-20s %5s %4s %5s\n", "Team", "W-L-O", "PTS", "+/-"))
	standingsTable.WriteString(strings.Repeat("-", 38) + "\n")
	for _, team := range divisionTeams {
		record := fmt.Sprintf("%d-%d-%d", team.Wins, team.Losses, team.OTLosses)
		marker := ""
		if team.TeamAbbrev.Default == match.TeamAbbrev.Default {
			marker = "▶"
		}
		standingsTable.WriteString(fmt.Sprintf("%s%-19s %5s %4d %+4d\n", marker, team.TeamName.Default, record, team.Points, team.GoalDifferential))
	}
	standingsTable.WriteString("```")

	embed := NewEmbed().
		SetTitle(match.TeamName.Default).
		SetDescription(fmt.Sprintf("%s Conference | %s Division", match.ConferenceName, match.DivisionName)).
		SetColor(ColorBlue).
		SetThumbnail(match.TeamLogo).
		AddField("Record", fmt.Sprintf("%d-%d-%d", match.Wins, match.Losses, match.OTLosses), true).
		AddField("Points", fmt.Sprintf("%d (%.3f)", match.Points, match.PointPctg), true).
		AddField("Goals", fmt.Sprintf("%d GF / %d GA (%+d)", match.GoalFor, match.GoalAgainst, match.GoalDifferential), true).
		AddField("Streak", fmt.Sprintf("%s%d", match.StreakCode, match.StreakCount), true).
		AddField(fmt.Sprintf("%s Division Standings", match.DivisionName), standingsTable.String(), false)

	return ctx.FollowUpEmbed(embed.Build())
}
