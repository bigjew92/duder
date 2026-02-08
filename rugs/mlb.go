package rugs

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&MLBCommand{})
}

// MLBCommand implements the /mlb slash command
type MLBCommand struct{}

func (c *MLBCommand) Name() string {
	return "mlb"
}

func (c *MLBCommand) Description() string {
	return "Lookup MLB player or team stats"
}

func (c *MLBCommand) Options() []*discordgo.ApplicationCommandOption {
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

func (c *MLBCommand) Execute(ctx CommandContext) error {
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

func (c *MLBCommand) handlePlayer(ctx CommandContext, playerName string) error {
	// Fetch all active MLB players
	currentYear := time.Now().Year()
	playersURL := fmt.Sprintf("https://statsapi.mlb.com/api/v1/sports/1/players?season=%d&gameType=R", currentYear)
	playersResp, err := ctx.HTTPGetString(10, playersURL, nil)
	if err != nil {
		return ctx.FollowUp("Failed to contact MLB API.")
	}

	type Player struct {
		ID       int    `json:"id"`
		FullName string `json:"fullName"`
		Team     struct {
			ID int `json:"id"`
		} `json:"currentTeam"`
		Position struct {
			Abbreviation string `json:"abbreviation"`
		} `json:"primaryPosition"`
	}

	var playersResult struct {
		People []Player `json:"people"`
	}

	if err := json.Unmarshal([]byte(playersResp), &playersResult); err != nil {
		return ctx.FollowUp("Failed to parse players list.")
	}

	// Find player by name (fuzzy match)
	nameLower := strings.ToLower(playerName)
	var player *Player
	for i := range playersResult.People {
		if strings.Contains(strings.ToLower(playersResult.People[i].FullName), nameLower) {
			player = &playersResult.People[i]
			break
		}
	}

	if player == nil {
		return ctx.FollowUp(fmt.Sprintf("No active MLB player found matching '%s'.", playerName))
	}

	// Fetch stats - determine if pitcher
	isPitcher := player.Position.Abbreviation == "P"
	statsGroup := "hitting"
	if isPitcher {
		statsGroup = "pitching"
	}

	statsURL := fmt.Sprintf("https://statsapi.mlb.com/api/v1/people/%d/stats?stats=season&season=%d&group=%s", player.ID, currentYear, statsGroup)
	statsResp, err := ctx.HTTPGetString(10, statsURL, nil)
	if err != nil {
		return ctx.FollowUp("Failed to fetch player stats.")
	}

	type StatSplit struct {
		Stat struct {
			Avg  string `json:"avg"`
			HR   int    `json:"homeRuns"`
			RBI  int    `json:"rbi"`
			OPS  string `json:"ops"`
			ERA  string `json:"era"`
			Wins int    `json:"wins"`
			Loss int    `json:"losses"`
			SO   int    `json:"strikeOuts"`
			WHIP string `json:"whip"`
		} `json:"stat"`
		Team struct {
			Name string `json:"name"`
		} `json:"team"`
	}

	var statsResult struct {
		Stats []struct {
			Splits []StatSplit `json:"splits"`
		} `json:"stats"`
	}

	json.Unmarshal([]byte(statsResp), &statsResult)

	// Get latest stats split
	var stats StatSplit
	var teamName string
	if len(statsResult.Stats) > 0 && len(statsResult.Stats[0].Splits) > 0 {
		stats = statsResult.Stats[0].Splits[len(statsResult.Stats[0].Splits)-1]
		teamName = stats.Team.Name
	}

	embed := NewEmbed().
		SetTitle(fmt.Sprintf("%s (%s)", player.FullName, player.Position.Abbreviation)).
		SetDescription(fmt.Sprintf("%s | Season %d", teamName, currentYear)).
		SetColor(ColorBlue).
		SetThumbnail(fmt.Sprintf("https://img.mlbstatic.com/mlb-photos/image/upload/d_people:generic:headshot:67:current.png/w_213,q_auto:best/v1/people/%d/headshot/67/current", player.ID))

	if isPitcher {
		embed.AddField("ERA", stats.Stat.ERA, true)
		embed.AddField("W-L", fmt.Sprintf("%d-%d", stats.Stat.Wins, stats.Stat.Loss), true)
		embed.AddField("SO", fmt.Sprintf("%d", stats.Stat.SO), true)
		embed.AddField("WHIP", stats.Stat.WHIP, true)
	} else {
		embed.AddField("AVG", stats.Stat.Avg, true)
		embed.AddField("HR", fmt.Sprintf("%d", stats.Stat.HR), true)
		embed.AddField("RBI", fmt.Sprintf("%d", stats.Stat.RBI), true)
		embed.AddField("OPS", stats.Stat.OPS, true)
	}

	return ctx.FollowUpEmbed(embed.Build())
}

func (c *MLBCommand) handleTeam(ctx CommandContext, teamName string) error {
	// MLB standings API - new statsapi endpoint
	currentYear := time.Now().Year()
	standingsURL := fmt.Sprintf("https://statsapi.mlb.com/api/v1/standings?leagueId=103,104&season=%d", currentYear)
	resp, err := ctx.HTTPGetString(10, standingsURL, nil)
	if err != nil {
		return ctx.FollowUp("Failed to fetch MLB standings.")
	}

	type TeamRecord struct {
		Team struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"team"`
		Wins              int    `json:"wins"`
		Losses            int    `json:"losses"`
		WinningPercentage string `json:"winningPercentage"`
		GamesBack         string `json:"gamesBack"`
		DivisionRank      string `json:"divisionRank"`
	}

	type DivisionRecord struct {
		Division struct {
			ID   int    `json:"id"`
			Name string `json:"name"`
		} `json:"division"`
		TeamRecords []TeamRecord `json:"teamRecords"`
	}

	var standingsResult struct {
		Records []DivisionRecord `json:"records"`
	}

	if err := json.Unmarshal([]byte(resp), &standingsResult); err != nil {
		return ctx.FollowUp("Failed to parse standings.")
	}

	// Find team across all divisions
	nameLower := strings.ToLower(teamName)
	var match *TeamRecord
	var matchDivision *DivisionRecord

	for i := range standingsResult.Records {
		div := &standingsResult.Records[i]
		for j := range div.TeamRecords {
			team := &div.TeamRecords[j]
			if strings.Contains(strings.ToLower(team.Team.Name), nameLower) {
				match = team
				matchDivision = div
				break
			}
		}
		if match != nil {
			break
		}
	}

	if match == nil {
		return ctx.FollowUp(fmt.Sprintf("No MLB team found matching '%s'.", teamName))
	}

	// Build division standings table
	var standingsTable strings.Builder
	standingsTable.WriteString("```\n")
	standingsTable.WriteString(fmt.Sprintf("%-22s %7s %5s %5s\n", "Team", "W-L", "PCT", "GB"))
	standingsTable.WriteString(strings.Repeat("-", 42) + "\n")
	for _, team := range matchDivision.TeamRecords {
		record := fmt.Sprintf("%d-%d", team.Wins, team.Losses)
		marker := ""
		if team.Team.ID == match.Team.ID {
			marker = "▶"
		}
		standingsTable.WriteString(fmt.Sprintf("%s%-21s %7s %5s %5s\n", marker, team.Team.Name, record, team.WinningPercentage, team.GamesBack))
	}
	standingsTable.WriteString("```")

	embed := NewEmbed().
		SetTitle(match.Team.Name).
		SetDescription(matchDivision.Division.Name).
		SetColor(ColorBlue).
		AddField("Record", fmt.Sprintf("%d-%d", match.Wins, match.Losses), true).
		AddField("Win %", match.WinningPercentage, true).
		AddField("GB", match.GamesBack, true).
		AddField(fmt.Sprintf("%s Standings", matchDivision.Division.Name), standingsTable.String(), false)

	return ctx.FollowUpEmbed(embed.Build())
}
