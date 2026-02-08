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
	// MLB standings API
	standingsURL := "http://lookup-service-prod.mlb.com/json/named.standings_schedule_date.bam?league_id='103','104'&season='2024'&stand_type='div'&schedule_game_date.game_date='2024-09-30'"
	resp, err := ctx.HTTPGetString(10, standingsURL, nil)
	if err != nil {
		return ctx.FollowUp("Failed to fetch MLB standings.")
	}

	type TeamStanding struct {
		TeamShort string `json:"team_short"`
		TeamFull  string `json:"team_full"`
		Wins      string `json:"w"`
		Losses    string `json:"l"`
		Pct       string `json:"pct"`
		GB        string `json:"gb"`
		Division  string `json:"division"`
	}

	var standingsResult struct {
		StandingsScheduleDate struct {
			StandingsAll struct {
				QueryResults struct {
					TotalSize string          `json:"totalSize"`
					Row       json.RawMessage `json:"row"`
				} `json:"queryResults"`
			} `json:"standings_all"`
		} `json:"standings_schedule_date"`
	}

	if err := json.Unmarshal([]byte(resp), &standingsResult); err != nil {
		return ctx.FollowUp("Failed to parse standings.")
	}

	totalSize, _ := strconv.Atoi(standingsResult.StandingsScheduleDate.StandingsAll.QueryResults.TotalSize)
	if totalSize == 0 {
		return ctx.FollowUp("No standings data available.")
	}

	var teams []TeamStanding
	if totalSize == 1 {
		var single TeamStanding
		json.Unmarshal(standingsResult.StandingsScheduleDate.StandingsAll.QueryResults.Row, &single)
		teams = []TeamStanding{single}
	} else {
		json.Unmarshal(standingsResult.StandingsScheduleDate.StandingsAll.QueryResults.Row, &teams)
	}

	// Find team
	nameLower := strings.ToLower(teamName)
	var match *TeamStanding

	for _, team := range teams {
		if strings.Contains(strings.ToLower(team.TeamFull), nameLower) ||
			strings.EqualFold(team.TeamShort, teamName) {
			match = &team
			break
		}
	}

	if match == nil {
		return ctx.FollowUp(fmt.Sprintf("No MLB team found matching '%s'.", teamName))
	}

	embed := NewEmbed().
		SetTitle(match.TeamFull).
		SetDescription(match.Division).
		SetColor(ColorBlue).
		AddField("Record", fmt.Sprintf("%s-%s", match.Wins, match.Losses), true).
		AddField("Win %", match.Pct, true).
		AddField("GB", match.GB, true)

	return ctx.FollowUpEmbed(embed.Build())
}
