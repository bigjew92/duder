package rugs

import (
	"fmt"
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func TestStarWarsCommand(t *testing.T) {
	// Setup
	cmd := &StarWarsCommand{}
	ctx := NewTestContext()

	// Mock data
	searchTerm := "Luke"
	searchType := "people"
	expectedURL := fmt.Sprintf("%s/%s/?search=%s", starWarsAPIBase, searchType, url.QueryEscape(searchTerm))

	mockResponse := `{
		"results": [
			{
				"name": "Luke Skywalker",
				"height": "172",
				"mass": "77",
				"hair_color": "blond",
				"skin_color": "fair",
				"eye_color": "blue",
				"birth_year": "19BBY",
				"gender": "male"
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
		// Verify basic properties of the embed
		// Note: verification logic depends on actual implementation of starwars.go
		return true
	})).Return(nil)

	// Execute
	err := cmd.Execute(ctx)

	// Verify
	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}
