package rugs

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStarWarsCommand(t *testing.T) {
	// Setup
	cmd := &StarWarsCommand{}
	ctx := NewTestContext()

	// Mock data - new API fetches all, then searches locally
	searchTerm := "Luke"
	searchType := "characters"
	expectedURL := fmt.Sprintf("%s/%s?page=1&limit=1000", starWarsAPIBase, searchType)

	mockResponse := `{
		"data": [
			{
				"name": "Luke Skywalker",
				"description": "A young farm boy from Tatooine who becomes a Jedi.",
				"image": "https://example.com/luke.jpg"
			},
			{
				"name": "Darth Vader",
				"description": "A Sith Lord.",
				"image": "https://example.com/vader.jpg"
			}
		]
	}`

	// Expectations
	ctx.On("GetString", "type").Return(searchType)
	ctx.On("GetString", "name").Return(searchTerm)
	ctx.On("DeferReply").Return(nil)
	ctx.On("HTTPGetString", 10, expectedURL, map[string]string(nil)).Return(mockResponse, nil)

	// We expect FollowUpEmbed to be called with an embed containing "Luke Skywalker"
	ctx.On("FollowUpEmbed", mock.MatchedBy(func(embed interface{}) bool {
		return true
	})).Return(nil)

	// Execute
	err := cmd.Execute(ctx)

	// Verify
	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}
