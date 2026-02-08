package rugs

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/bwmarrin/discordgo"
)

func init() {
	Register(&DiceCommand{})
	Register(&EightBallCommand{})
	Register(&LebowskiCommand{})
	Register(&DadJokeCommand{})
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

// DadJokeCommand implements the /dadjoke slash command
type DadJokeCommand struct{}

func (c *DadJokeCommand) Name() string {
	return "dadjoke"
}

func (c *DadJokeCommand) Description() string {
	return "Get a random dad joke"
}

func (c *DadJokeCommand) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

func (c *DadJokeCommand) Execute(ctx CommandContext) error {
	ctx.DeferReply()

	headers := map[string]string{
		"Accept":     "text/plain",
		"User-Agent": "Duder/1.0 (https://github.com/bigjew92/duder)",
	}

	resp, err := ctx.HTTPGetString(10, "https://icanhazdadjoke.com/", headers)
	if err != nil {
		return ctx.FollowUp("Failed to fetch dad joke.")
	}

	return ctx.FollowUp(fmt.Sprintf("```%s```", resp))
}

// BigCommand implements the /big slash command
type BigCommand struct{}

var bigLetters = map[rune]string{
	' ': " ",
	'0': ":zero:", '1': ":one:", '2': ":two:", '3': ":three:", '4': ":four:",
	'5': ":five:", '6': ":six:", '7': ":seven:", '8': ":eight:", '9': ":nine:",
	'!': ":exclamation:", '?': ":question:",
	'a': ":regional_indicator_a:", 'b': ":regional_indicator_b:", 'c': ":regional_indicator_c:",
	'd': ":regional_indicator_d:", 'e': ":regional_indicator_e:", 'f': ":regional_indicator_f:",
	'g': ":regional_indicator_g:", 'h': ":regional_indicator_h:", 'i': ":regional_indicator_i:",
	'j': ":regional_indicator_j:", 'k': ":regional_indicator_k:", 'l': ":regional_indicator_l:",
	'm': ":regional_indicator_m:", 'n': ":regional_indicator_n:", 'o': ":regional_indicator_o:",
	'p': ":regional_indicator_p:", 'q': ":regional_indicator_q:", 'r': ":regional_indicator_r:",
	's': ":regional_indicator_s:", 't': ":regional_indicator_t:", 'u': ":regional_indicator_u:",
	'v': ":regional_indicator_v:", 'w': ":regional_indicator_w:", 'x': ":regional_indicator_x:",
	'y': ":regional_indicator_y:", 'z': ":regional_indicator_z:",
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
	return "Get a random Aura copypasta quote"
}

func (c *AuraCommand) Options() []*discordgo.ApplicationCommandOption {
	return nil
}

var auraResponses = []string{
	"Hello ora and chat This is the governor of Alabama, Al Bama. I regret to inform you that our scientists have gone extinct. One of the Alabama Beach Mice accidentally put Him in the microwave since he thought it was Jimmy Changa. Our bad. We have sent Ora a statue of the beloved scientist to put on his front lawn to commemorate his life. We know Ora loved the scientis as much as the Alabama Beach Mouse. We hope to see the scientist prominently displayed on his front lawn.",
	"To those who think lose = boo, what you of lesser intellect don't understand is Aurateur is such a titan of Mario Maker that he is literally guaranteed to win on every single good level. Since he only likes good levels, he wins only on levels he likes. When he boos, it's because he has deciphered every fiber that composes the level, meticulously figuring out the electron orientations that make it bad. Aura is a god of Mario, and when his boo hammer strikes, it brings justice!",
	"Hi Aurateur- It’s me, your only viewer. For years I have created the illusion that you’re streaming to a large audience. The truth? All these people in the chat are me. And now, to convince you, I will send this message from all my accounts",
	"Idiots. You fucking morons. What are you doing. Spawn here. Toad, stay here. No, what are you doing, you fucking fool. Ugh, you guys are all such oafs. Okay, let’s try this again. Toadette?! What are you doing? Why did you run off? You fucking nincompoop? Ugh, these clowns have no idea what they’re doing. Okay, pick me up Mario and throw me to that ledge. What was that?! You fucking dope. Don’t throw me at the wall. The ground there. Ugh, these imbeciles have no idea what they’re doing",
	"Look at this chat, it's ridiculous, one of the worst things i've ever seen, the chat does not respect anything or anyone, the only thing they can do is copy and paste the same thing over and over again, and not only that some dudes actually pay money to a bald man so they can make this nice girl named Joanna to say some dumb things that represents their stupidity as a human being. Please stop with this, no one likes it, not even the Streamer, or Joanna.",
	"hey aura. so. i got this new anime plot. basically there's this high school girl except she's got huge boobs. i mean some serious honkers. a real set of badonkers. packin some dobonhonkeros. massive dohoonkabhankoloos. big ol' tonhongerekoogers. what happens next?! transfer student shows up with even bigger bonkhonagahoogs. humongous hungolomghononoloughongous. so what do you think aura? pretty good right? hmm?",
	"Licky Ricky bald bald lucky licky Ricky, licky licky bald bald licky ricky bald. Licky Ricky bald bald lucky licky Ricky, licky licky bald bald licky ricky bald. Licky Ricky bald bald lucky licky Ricky, licky licky bald bald licky ricky bald. Licky Ricky bald bald lucky licky Ricky, licky licky bald bald licky ricky bald. Licky Ricky bald bald lucky licky Ricky, licky licky bald bald licky ricky bald. Licky Ricky bald bald lucky licky Ricky, licky licky bald bald licky ricky bald.",
	"So Aurateur, let's have a talk. I strongly feel that your channel would do better if you turned off TTS or at least filtered out the spam. Thing is, while I myself have never donated to you or any other streamer, and my only subs are twitch prime, I feel I know a thing or two about how to succeed as a streamer and I'm certain your channel would do better without all the spam, it's clear nobody wants that shit and is the only thing holding you back as a fulltime streamer. Thank you for listening.",
	"Toad what are you doing. Mario, die. Just die. All of you die. Now, now, spawn on me. Ok, here's what we're gonna do. Mario.. wait a second... Actually, toad, YOU pick up Mario. I'm going to jump off this edge, and you take Mario and jump off a little bit after, bounce off the top of my head, and then just throw him....... OMG, Toad, where were you? You were supposed to... Ok, were going to try this again... Now, listen close as I explain it to you one more time.",
	"The Alabama beach mouse (Peromyscus polionotus ammobates) is a federally endangered species which lives along the Alabama coast. The range of the Alabama beach mouse historically included much of the Fort Morgan Peninsula on the Alabama Gulf coast and extends from Ono Island to Fort Morgan. Coastal residential and commercial development and roadway construction have fragmented and destroyed habitat used by this species. Hurricanes, tropical storms.",
}

func (c *AuraCommand) Execute(ctx CommandContext) error {
	response := auraResponses[RandomInRange(0, len(auraResponses)-1)]
	return ctx.Reply(response)
}

// Helper function for int pointer
func intPtr(i int) *float64 {
	f := float64(i)
	return &f
}
