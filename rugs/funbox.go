package rugs

import (
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&DiceCommand{})
	Register(&EightBallCommand{})
	Register(&LebowskiCommand{})
	Register(&BashCommand{})
	Register(&BigCommand{})
	Register(&SmolCommand{})
	Register(&AuraCommand{})
}

// DiceCommand implements the /dice slash command
type DiceCommand struct{}

func (c *DiceCommand) Name() string {
	return "dice"
}

func (c *DiceCommand) Description() string {
	return "Roll a dice"
}

func (c *DiceCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionInteger,
			Name:        "sides",
			Description: "Number of sides (default: 6)",
			Required:    false,
			MinValue:    intPtr(2),
			MaxValue:    99,
		},
	}
}

func (c *DiceCommand) Execute(ctx CommandContext) error {
	sides := int(ctx.GetInt("sides"))
	if sides == 0 {
		sides = 6
	}
	sides = Clamp(sides, 2, 99)

	roll := RandomInRange(1, sides)
	return ctx.Reply(fmt.Sprintf("🎲 Rolled a %d-sided dice and got **%d**!", sides, roll))
}

// EightBallCommand implements the /8ball slash command
type EightBallCommand struct{}

var eightBallResponses = []string{
	"It is certain",
	"It is decidedly so",
	"Without a doubt",
	"Yes, definitely",
	"You may rely on it",
	"As I see it, yes",
	"Most likely",
	"Outlook good",
	"Yes",
	"Signs point to yes",
	"Reply hazy, try again",
	"Ask again later",
	"Better not tell you now",
	"Cannot predict now",
	"Concentrate and ask again",
	"Don't count on it",
	"My reply is no",
	"My sources say no",
	"Outlook not so good",
	"Very doubtful",
}

func (c *EightBallCommand) Name() string {
	return "8ball"
}

func (c *EightBallCommand) Description() string {
	return "Ask the magic 8-ball a question"
}

func (c *EightBallCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "question",
			Description: "Your question for the 8-ball",
			Required:    false,
		},
	}
}

func (c *EightBallCommand) Execute(ctx CommandContext) error {
	response := eightBallResponses[RandomInRange(0, len(eightBallResponses)-1)]
	return ctx.Reply(fmt.Sprintf("🎱 %s", response))
}

// LebowskiCommand implements the /lebowski slash command
type LebowskiCommand struct{}

func (c *LebowskiCommand) Name() string {
	return "lebowski"
}

func (c *LebowskiCommand) Description() string {
	return "Get a random Big Lebowski quote"
}

func (c *LebowskiCommand) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

func (c *LebowskiCommand) Execute(ctx CommandContext) error {
	ctx.DeferReply()

	resp, err := ctx.HTTPGetString(10, "https://lebowski.me/api/quotes/random", nil)
	if err != nil {
		return ctx.FollowUp("The Dude could not be reached. Try again later.")
	}

	var result struct {
		Quote struct {
			Content string `json:"content"`
		} `json:"quote"`
	}
	if err := json.Unmarshal([]byte(resp), &result); err != nil {
		return ctx.FollowUp("Failed to parse quote.")
	}

	return ctx.FollowUp(fmt.Sprintf("```%s```", result.Quote.Content))
}

// BashCommand implements the /bash slash command
type BashCommand struct{}

func (c *BashCommand) Name() string {
	return "bash"
}

func (c *BashCommand) Description() string {
	return "Get a random bash.org quote"
}

func (c *BashCommand) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

func (c *BashCommand) Execute(ctx CommandContext) error {
	ctx.DeferReply()

	resp, err := ctx.HTTPGetString(10, "http://bash.org/?random", nil)
	if err != nil {
		return ctx.FollowUp("Failed to fetch bash.org quote.")
	}

	// Parse bash.org HTML (simplified)
	re := regexp.MustCompile(`<p class="qt">(.*?)</p>`)
	matches := re.FindStringSubmatch(resp)
	if len(matches) < 2 {
		return ctx.FollowUp("Failed to parse bash.org quote.")
	}

	quote := DecodeHTML(matches[1])
	quote = strings.ReplaceAll(quote, "<br />", "\n")

	if len(quote) > 1900 {
		quote = quote[:1900] + "..."
	}

	return ctx.FollowUp(fmt.Sprintf("```%s```", quote))
}

// BigCommand implements the /big slash command
type BigCommand struct{}

var bigLetters = map[rune]string{
	'a': "🇦", 'b': "🇧", 'c': "🇨", 'd': "🇩", 'e': "🇪",
	'f': "🇫", 'g': "🇬", 'h': "🇭", 'i': "🇮", 'j': "🇯",
	'k': "🇰", 'l': "🇱", 'm': "🇲", 'n': "🇳", 'o': "🇴",
	'p': "🇵", 'q': "🇶", 'r': "🇷", 's': "🇸", 't': "🇹",
	'u': "🇺", 'v': "🇻", 'w': "🇼", 'x': "🇽", 'y': "🇾",
	'z': "🇿", ' ': "   ",
	'0': "0️⃣", '1': "1️⃣", '2': "2️⃣", '3': "3️⃣", '4': "4️⃣",
	'5': "5️⃣", '6': "6️⃣", '7': "7️⃣", '8': "8️⃣", '9': "9️⃣",
}

func (c *BigCommand) Name() string {
	return "big"
}

func (c *BigCommand) Description() string {
	return "Convert text to big emoji letters"
}

func (c *BigCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "text",
			Description: "Text to convert",
			Required:    true,
		},
	}
}

func (c *BigCommand) Execute(ctx CommandContext) error {
	text := strings.ToLower(ctx.GetString("text"))

	var result strings.Builder
	for _, char := range text {
		if emoji, ok := bigLetters[char]; ok {
			result.WriteString(emoji)
		} else {
			result.WriteRune(char)
		}
	}

	output := result.String()
	if len(output) > 2000 {
		return ctx.ReplyEphemeral("Text too long!")
	}

	return ctx.Reply(output)
}

// SmolCommand implements the /smol slash command
type SmolCommand struct{}

var smolLetters = map[rune]rune{
	'a': 'ᵃ', 'b': 'ᵇ', 'c': 'ᶜ', 'd': 'ᵈ', 'e': 'ᵉ',
	'f': 'ᶠ', 'g': 'ᵍ', 'h': 'ʰ', 'i': 'ⁱ', 'j': 'ʲ',
	'k': 'ᵏ', 'l': 'ˡ', 'm': 'ᵐ', 'n': 'ⁿ', 'o': 'ᵒ',
	'p': 'ᵖ', 'q': 'q', 'r': 'ʳ', 's': 'ˢ', 't': 'ᵗ',
	'u': 'ᵘ', 'v': 'ᵛ', 'w': 'ʷ', 'x': 'ˣ', 'y': 'ʸ',
	'z': 'ᶻ',
}

func (c *SmolCommand) Name() string {
	return "smol"
}

func (c *SmolCommand) Description() string {
	return "Convert text to tiny superscript letters"
}

func (c *SmolCommand) Options() []*discordgo.ApplicationCommandOption {
	return []*discordgo.ApplicationCommandOption{
		{
			Type:        discordgo.ApplicationCommandOptionString,
			Name:        "text",
			Description: "Text to convert",
			Required:    true,
		},
	}
}

func (c *SmolCommand) Execute(ctx CommandContext) error {
	text := strings.ToLower(ctx.GetString("text"))

	var result strings.Builder
	for _, char := range text {
		if smol, ok := smolLetters[char]; ok {
			result.WriteRune(smol)
		} else {
			result.WriteRune(char)
		}
	}

	return ctx.Reply(result.String())
}

// AuraCommand implements the /aura slash command
type AuraCommand struct{}

func (c *AuraCommand) Name() string {
	return "aura"
}

func (c *AuraCommand) Description() string {
	return "Check your aura level"
}

func (c *AuraCommand) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

func (c *AuraCommand) Execute(ctx CommandContext) error {
	aura := RandomInRange(-1000, 1000)

	var emoji string
	var desc string
	switch {
	case aura >= 800:
		emoji = "✨"
		desc = "Legendary aura! You radiate pure energy!"
	case aura >= 500:
		emoji = "🌟"
		desc = "Amazing aura! People are drawn to you!"
	case aura >= 200:
		emoji = "😊"
		desc = "Good vibes! Your aura is positive!"
	case aura >= 0:
		emoji = "😐"
		desc = "Neutral aura. Nothing special today."
	case aura >= -200:
		emoji = "😕"
		desc = "Slightly negative aura. Watch out!"
	case aura >= -500:
		emoji = "😰"
		desc = "Bad aura detected. Stay inside maybe?"
	default:
		emoji = "💀"
		desc = "Catastrophic aura levels. What did you do?!"
	}

	sign := ""
	if aura >= 0 {
		sign = "+"
	}

	return ctx.Reply(fmt.Sprintf("%s **%s%d aura**\n%s", emoji, sign, aura, desc))
}

// Helper function for int pointer
func intPtr(i int) *float64 {
	f := float64(i)
	return &f
}
