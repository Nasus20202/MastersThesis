package inference

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestToolCallReplayArguments(t *testing.T) {
	t.Parallel()

	assert.Equal(t, `{"command":"true"}`, ToolCall{Arguments: `{"command":"true"}`}.ReplayArguments())
	assert.Equal(t, "{}", ToolCall{Arguments: `{"command":"kubectl get`}.ReplayArguments())
	assert.Equal(t, "{}", ToolCall{Arguments: ""}.ReplayArguments())
}
