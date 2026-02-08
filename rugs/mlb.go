package rugs

import (
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
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
	// Search for player (append %25 for partial match)
	searchURL := fmt.Sprintf("http://lookup-service-prod.mlb.com/json/named.search_player_all.bam?sport_code='mlb'&active_sw='Y'&name_part='%s%%25'", url.QueryEscape(playerName))
	searchResp, err := ctx.HTTPGetString(10, searchURL, nil)
	if err != nil {
		return ctx.FollowUp("Failed to contact MLB API.")
	}

	type PlayerRow struct {
		PlayerID string `json:"player_id"`
		Name     string `json:"name_display_first_last"`
		Team     string `json:"team_full"`
		Position string `json:"position"`
	}

	var searchResult struct {
		SearchPlayerAll struct {
			QueryResults struct {
				TotalSize string          `json:"totalSize"`
				Row       json.RawMessage `json:"row"`
			} `json:"queryResults"`
		} `json:"search_player_all"`
	}

	if err := json.Unmarshal([]byte(searchResp), &searchResult); err != nil {
		return ctx.FollowUp("Failed to parse search results.")
	}

	totalSize, _ := strconv.Atoi(searchResult.SearchPlayerAll.QueryResults.TotalSize)
	if totalSize == 0 {
		return ctx.FollowUp("No active MLB players found.")
	}

	var player PlayerRow
	if totalSize == 1 {
		json.Unmarshal(searchResult.SearchPlayerAll.QueryResults.Row, &player)
	} else {
		var players []PlayerRow
		json.Unmarshal(searchResult.SearchPlayerAll.QueryResults.Row, &players)
		player = players[0]
	}

	// Fetch stats
	isPitcher := player.Position == "P"
	currentYear := time.Now().Year()

	type StatsRow struct {
		Avg  string `json:"avg"`
		HR   string `json:"hr"`
		RBI  string `json:"rbi"`
		OPS  string `json:"ops"`
		ERA  string `json:"era"`
		Wins string `json:"w"`
		Loss string `json:"l"`
		SO   string `json:"so"`
		WHIP string `json:"whip"`
	}

	var statsURL string
	if isPitcher {
		statsURL = fmt.Sprintf("http://lookup-service-prod.mlb.com/json/named.sport_pitching_tm.bam?league_list_id='mlb'&game_type='R'&season='%d'&player_id='%s'", currentYear, player.PlayerID)
	} else {
		statsURL = fmt.Sprintf("http://lookup-service-prod.mlb.com/json/named.sport_hitting_tm.bam?league_list_id='mlb'&game_type='R'&season='%d'&player_id='%s'", currentYear, player.PlayerID)
	}

	statsResp, err := ctx.HTTPGetString(10, statsURL, nil)
	if err != nil {
		return ctx.FollowUp("Failed to fetch stats.")
	}

	var rawMap map[string]json.RawMessage
	json.Unmarshal([]byte(statsResp), &rawMap)

	var innerJSON json.RawMessage
	if isPitcher {
		innerJSON = rawMap["sport_pitching_tm"]
	} else {
		innerJSON = rawMap["sport_hitting_tm"]
	}

	var statsResult struct {
		QueryResults struct {
			TotalSize string          `json:"totalSize"`
			Row       json.RawMessage `json:"row"`
		} `json:"queryResults"`
	}

	var stats StatsRow
	if innerJSON != nil {
		json.Unmarshal(innerJSON, &statsResult)
		totalStats, _ := strconv.Atoi(statsResult.QueryResults.TotalSize)
		if totalStats > 0 {
			if totalStats == 1 {
				json.Unmarshal(statsResult.QueryResults.Row, &stats)
			} else {
				var multipleStats []StatsRow
				json.Unmarshal(statsResult.QueryResults.Row, &multipleStats)
				stats = multipleStats[len(multipleStats)-1]
			}
		}
	}

	embed := NewEmbed().
		SetTitle(fmt.Sprintf("%s (%s)", player.Name, player.Position)).
		SetDescription(fmt.Sprintf("%s | Season %d", player.Team, currentYear)).
		SetColor(ColorBlue).
		SetThumbnail(fmt.Sprintf("https://img.mlbstatic.com/mlb-photos/image/upload/d_people:generic:headshot:67:current.png/w_213,q_auto:best/v1/people/%s/headshot/67/current", player.PlayerID))

	if isPitcher {
		embed.AddField("ERA", stats.ERA, true)
		embed.AddField("W-L", fmt.Sprintf("%s-%s", stats.Wins, stats.Loss), true)
		embed.AddField("SO", stats.SO, true)
		embed.AddField("WHIP", stats.WHIP, true)
	} else {
		embed.AddField("AVG", stats.Avg, true)
		embed.AddField("HR", stats.HR, true)
		embed.AddField("RBI", stats.RBI, true)
		embed.AddField("OPS", stats.OPS, true)
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
