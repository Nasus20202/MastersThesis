package analysis

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMeanOverModelsWeighsModelsEquallyAndPairsByModel(t *testing.T) {
	attempts := []RunAttempt{
		{Model: "big", Scenario: "a", Score: 1},
		{Model: "big", Scenario: "b", Score: 1},
		{Model: "small", Scenario: "a", Score: 0.5},
		{Model: "small", Scenario: "a", Score: 0.5},
		{Model: "small", Scenario: "b", Score: 0},
		{Model: "solo", Scenario: "a", Score: 0},
	}
	reference := []RunAttempt{
		{Model: "big", Scenario: "a", Score: 0},
		{Model: "big", Scenario: "b", Score: 1},
		{Model: "small", Scenario: "a", Score: 0.5},
		{Model: "small", Scenario: "b", Score: 0.5},
	}

	models, groups := ByModel(attempts)
	assert.Equal(t, []string{"big", "small", "solo"}, models)
	assert.Len(t, groups["small"], 3)

	mean := MeanOverModels(attempts, reference)
	assert.Equal(t, []string{"big", "small", "solo"}, mean.Models)
	// big 1.0, small 0.25, solo 0.0
	assert.InDelta(t, 1.25/3, mean.Macro, 1e-9)
	require.NotNil(t, mean.ReferenceMacro)
	// big 0.5, small 0.5; solo has no reference
	assert.InDelta(t, 0.5, *mean.ReferenceMacro, 1e-9)
	assert.Equal(t, map[string]float64{"big": 0.5, "small": 0.5}, mean.ReferenceMacros)
	require.NotNil(t, mean.Difference)
	// scenario a: big +1, small 0 -> +0.5; scenario b: big 0, small -0.5 -> -0.25
	assert.Equal(t, 2, mean.Difference.Scenarios)
	assert.InDelta(t, 0.125, mean.Difference.Mean, 1e-9)
}

func TestMeanOverModelsWithoutReference(t *testing.T) {
	mean := MeanOverModels([]RunAttempt{{Model: "m", Scenario: "a", Score: 1}}, nil)
	assert.InDelta(t, 1, mean.Macro, 1e-9)
	assert.Nil(t, mean.ReferenceMacro)
	assert.Nil(t, mean.Difference)
}
