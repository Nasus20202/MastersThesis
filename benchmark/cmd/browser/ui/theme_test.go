package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
)

func TestTerminationStyles(t *testing.T) {
	assert.Equal(t, Success, Termination(common.TerminationCompleted))
	assert.Equal(t, Danger, Termination(common.TerminationInference))
	assert.Equal(t, Warning, Termination(common.TerminationTurnLimit))
}
