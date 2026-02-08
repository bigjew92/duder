package rugs

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func Test8BallCommand(t *testing.T) {
	// Setup
	cmd := &EightBallCommand{}
	ctx := NewTestContext()

	// 8ball replies with a random string. We match any string.
	ctx.On("Reply", mock.AnythingOfType("string")).Return(nil)

	// Execute
	err := cmd.Execute(ctx)

	// Verify
	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}

func TestDiceCommand(t *testing.T) {
	// Setup
	cmd := &DiceCommand{}
	ctx := NewTestContext()

	// Expectations
	// Mock getting the "sides" option. Return 0 to trigger default 6 sides.
	ctx.On("GetInt", "sides").Return(int64(0))

	// Expect a reply containing "Rolled a 6-sided dice"
	ctx.On("Reply", mock.MatchedBy(func(content string) bool {
		return len(content) > 0 // We can't predict the random number, but it should return a string
	})).Return(nil)

	// Execute
	err := cmd.Execute(ctx)

	// Verify
	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}

func TestLebowskiCommand(t *testing.T) {
	cmd := &LebowskiCommand{}
	ctx := NewTestContext()

	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, "https://lebowski.me/api/quotes/random", map[string]string(nil)).
		Return(`{"quote":{"content":"The Dude abides."}}`, nil)
	ctx.On("FollowUp", "```The Dude abides.```").Return(nil)

	err := cmd.Execute(ctx)
	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}

func TestBashCommand(t *testing.T) {
	cmd := &BashCommand{}
	ctx := NewTestContext()

	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, "http://bash.org/?random", map[string]string(nil)).
		Return(`<p class="qt">I put on my robe and wizard hat.</p>`, nil)
	ctx.On("FollowUp", "```I put on my robe and wizard hat.```").Return(nil)

	err := cmd.Execute(ctx)
	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}

func TestBigCommand(t *testing.T) {
	cmd := &BigCommand{}
	ctx := NewTestContext()

	ctx.On("GetString", "text").Return("abc")
	ctx.On("Reply", "🇦🇧🇨").Return(nil)

	err := cmd.Execute(ctx)
	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}

func TestSmolCommand(t *testing.T) {
	cmd := &SmolCommand{}
	ctx := NewTestContext()

	ctx.On("GetString", "text").Return("ABC")
	ctx.On("Reply", "ᵃᵇᶜ").Return(nil)

	err := cmd.Execute(ctx)
	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}

func TestAuraCommand(t *testing.T) {
	cmd := &AuraCommand{}
	ctx := NewTestContext()

	// Aura replies with a random string. We match any string.
	ctx.On("Reply", mock.AnythingOfType("string")).Return(nil)

	err := cmd.Execute(ctx)
	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}
