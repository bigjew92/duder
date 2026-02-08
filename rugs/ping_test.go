package rugs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPingCommand(t *testing.T) {
	// Setup
	cmd := &PingCommand{}
	ctx := NewTestContext()

	// Expectations
	ctx.On("Reply", "pong 🏓").Return(nil)

	// Execute
	err := cmd.Execute(ctx)

	// Verify
	assert.NoError(t, err)
	ctx.AssertExpectations(t)
}
