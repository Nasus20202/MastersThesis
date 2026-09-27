package screens

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/model"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

func TestRunsPreviewComparesOnlyTheScopedRuns(t *testing.T) {
	root := t.TempDir()
	for _, run := range []struct{ id, agent string }{{"run-skill", "skill"}, {"run-baseline", "baseline"}} {
		store, err := results.New(root, results.RunMetadata{RunID: run.id, Agents: []string{run.agent}, Scenarios: []string{"alpha"}, Parallelism: 1, RepeatCount: 1})
		require.NoError(t, err)
		require.NoError(t, store.WriteAttempt(1, run.agent, orchestration.RunResult{
			ScenarioID: "alpha", Condition: run.agent,
			Agent:   &common.Result{Termination: common.TerminationCompleted},
			Grading: orchestration.GradingResult{Score: 1, FullSuccess: true},
		}))
	}

	read := model.NewStore(model.StoreConfig{ResultsRoot: root})
	require.NoError(t, read.Reload())
	view := NewView(read)

	out := view.runsPreview(&Route{Kind: AgentRuns, Agent: "skill"}, 120)
	assert.Contains(t, out, "run-skill")
	assert.NotContains(t, out, "run-baseline")
}
