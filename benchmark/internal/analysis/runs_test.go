package analysis

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChangesCluster(t *testing.T) {
	for command, want := range map[string]bool{
		"kubectl get pods -A":                                                     false,
		"kubectl describe deployment app":                                         false,
		"kubectl rollout status deployment/app":                                   false,
		"kubectl patch deployment app -p '{}'":                                    true,
		"kubectl -n app set image deployment/app app=nginx":                       true,
		"kubectl rollout restart deployment app":                                  true,
		"cat <<EOF | kubectl apply -f -":                                          true,
		"kubectl apply --dry-run=server -f fix.yaml":                              false,
		"kubectl apply --dry-run=server -f fix.yaml && kubectl apply -f fix.yaml": true,
		"kubectl get pods | grep delete":                                          false,
		"kubectl auth can-i create pods --as system:anything":                     false,
	} {
		assert.Equal(t, want, ChangesCluster(command), command)
	}
}

func TestLoadSources(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "networking", "service-port")
	require.NoError(t, os.MkdirAll(dir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "source.md"), []byte("- Source path: `content/en/docs/service.md`\n"), 0o600))

	sources, err := LoadSources(root)

	require.NoError(t, err)
	assert.Equal(t, map[string]string{"service-port": "content/en/docs/service.md"}, sources)
}

func TestLoadSubset(t *testing.T) {
	path := filepath.Join(t.TempDir(), "subset.txt")
	require.NoError(t, os.WriteFile(path, []byte("scenarios/networking/service-port\nscenarios/storage/pvc-pending\n"), 0o600))

	include, err := LoadSubset(path)

	require.NoError(t, err)
	assert.True(t, include("service-port"))
	assert.True(t, include("pvc-pending"))
	assert.False(t, include("other"))

	all, err := LoadSubset("")
	require.NoError(t, err)
	assert.True(t, all("other"))
}

func TestNewRunAttemptOrdersSearchesAgainstFirstChange(t *testing.T) {
	hits := func(paths ...string) map[string]any {
		var list []any
		for index, path := range paths {
			list = append(list, map[string]any{"rank": index + 1, "chunk": map[string]any{"path": path}})
		}
		return map[string]any{"query": "service port", "hits": list}
	}
	result := results.AttemptResult{
		Condition:  "rag",
		ScenarioID: "service-port",
		Attempt:    2,
		Grading:    orchestration.GradingResult{Score: 1, FullSuccess: true},
		Agent: &common.Result{Termination: "completed", Inference: inference.Metadata{Model: "gemma"}, ContextOverflow: true, TokenUsage: common.TokenUsage{PeakContextTokens: 31000}, ToolCalls: []common.ToolCallEvidence{
			{Call: inference.ToolCall{Name: "bash", Arguments: `{"command":"kubectl get svc"}`}},
			{Call: inference.ToolCall{Name: SearchToolName}, Details: hits("other.md", "service.md")},
			{Call: inference.ToolCall{Name: "bash", Arguments: `{"command":"kubectl patch svc app -p '{}'"}`}},
			{Call: inference.ToolCall{Name: SearchToolName}, Details: hits("other.md"), Error: "search failed"},
		}},
	}

	attempt, err := newRunAttempt("run", result, map[string]string{"service-port": "service.md"})

	require.NoError(t, err)
	assert.True(t, attempt.AgentRan)
	assert.Equal(t, 2, attempt.FirstChange)
	require.Len(t, attempt.Searches, 2)
	assert.Equal(t, Search{Call: 1, Query: "service port", BeforeChange: true, SourceRank: 2, Paths: []string{"other.md", "service.md"}}, attempt.Searches[0])
	assert.False(t, attempt.Searches[1].BeforeChange)
	assert.Zero(t, attempt.Searches[1].SourceRank)
	assert.Equal(t, "search failed", attempt.Searches[1].Error)
	assert.True(t, attempt.SourceRetrieved())

	_, err = newRunAttempt("run", result, map[string]string{})
	assert.Error(t, err)
}

func TestNewRunAttemptCountsChangesAndChecks(t *testing.T) {
	bash := func(command string, exitCode int) common.ToolCallEvidence {
		arguments, _ := json.Marshal(map[string]string{"command": command})
		return common.ToolCallEvidence{Call: inference.ToolCall{Name: "bash", Arguments: string(arguments)}, Details: map[string]any{"command": command, "exit_code": exitCode}}
	}
	result := results.AttemptResult{
		ScenarioID: "app",
		Grading:    orchestration.GradingResult{Score: 0.5},
		Agent: &common.Result{Termination: "completed", Inference: inference.Metadata{Model: "gemma"}, ContextOverflow: true, TokenUsage: common.TokenUsage{PeakContextTokens: 31000}, ToolCalls: []common.ToolCallEvidence{
			bash("kubectl edit deployment app", 1),
			bash("kubectl patch deployment app --type merge -p '{}'", 1),
			bash("kubectl set image deployment/app app=nginx && kubectl rollout status deployment/app --timeout=50s", 0),
		}},
	}

	attempt, err := newRunAttempt("run", result, map[string]string{"app": "app.md"})
	require.NoError(t, err)
	assert.Equal(t, 0, attempt.FirstChange)
	assert.Equal(t, 3, attempt.Changes)
	assert.Equal(t, 2, attempt.FailedChanges)
	assert.Equal(t, 1, attempt.Edits)
	assert.True(t, attempt.RolloutStatus)
	assert.Equal(t, "gemma", attempt.Model)
	assert.Equal(t, 31000, attempt.PeakContext)
	assert.True(t, attempt.Overflow)

	usage := Summarize("prompt", []RunAttempt{attempt}, nil).Usage
	assert.Equal(t, Usage{Attempts: 1, PeakContext: 31000, MaxPeakContext: 31000, Overflows: 1, Changes: 3, FailedChanges: 2, Edits: 1, RolloutStatus: 1, Unconfirmed: 1}, usage)
}

func TestSummarizeGroupsAttemptsAndMatchesReferenceScenarios(t *testing.T) {
	found := []Search{{SourceRank: 1, BeforeChange: true}}
	missed := []Search{{}}
	attempts := []RunAttempt{
		{Scenario: "a", Score: 1, FullSuccess: true, Searches: found},
		{Scenario: "a", Score: 0.5, Searches: missed},
		{Scenario: "b", Score: 0},
	}
	attempts[0].AgentRan, attempts[1].AgentRan = true, true
	reference := []RunAttempt{
		{Scenario: "a", Score: 0.5},
		{Scenario: "a", Score: 0.5},
		{Scenario: "b", Score: 1, FullSuccess: true},
	}

	summary := Summarize("rag", attempts, reference)

	assert.InDelta(t, 0.375, summary.Macro, 1e-9)
	assert.Equal(t, 1, summary.FullSuccess)
	assert.Equal(t, 1, summary.AgentNotRun)
	assert.Equal(t, 2, summary.SearchAttempts)
	assert.Equal(t, 2, summary.Searches)
	assert.Equal(t, 1, summary.SearchesBeforeChange)
	assert.Equal(t, 1, summary.SourceHits)
	assert.Equal(t, 1, summary.SourceAttempts)
	require.Len(t, summary.Groups, 3)
	noSearch, missedGroup, foundGroup := summary.Groups[0], summary.Groups[1], summary.Groups[2]
	assert.Equal(t, OutcomeGroup{Name: GroupNoSearch, Attempts: 1, Macro: 0, Scenarios: 1, ReferenceMacro: 1, ReferenceFull: 1, ReferenceAttempts: 1, ReferenceAvailable: true}, noSearch)
	assert.InDelta(t, 0.5, missedGroup.Macro, 1e-9)
	assert.InDelta(t, 0.5, missedGroup.ReferenceMacro, 1e-9)
	assert.Equal(t, 2, missedGroup.ReferenceAttempts)
	assert.InDelta(t, 1, foundGroup.Macro, 1e-9)
	require.NotNil(t, summary.MacroDifference)
	assert.Equal(t, 2, summary.MacroDifference.Scenarios)
	assert.InDelta(t, -0.375, summary.MacroDifference.Mean, 1e-9)
	assert.InDelta(t, -1, summary.MacroDifference.Low, 1e-9)
	assert.InDelta(t, 0.25, summary.MacroDifference.High, 1e-9)
	require.Len(t, summary.Scenarios, 2)
	assert.Equal(t, "a", summary.Scenarios[0].Scenario)
	assert.InDelta(t, 0.75, summary.Scenarios[0].MeanScore, 1e-9)
	assert.Equal(t, 1, summary.Scenarios[0].SourceRetrieved)
	require.NotNil(t, summary.Scenarios[0].ReferenceScore)
	assert.InDelta(t, 0.5, *summary.Scenarios[0].ReferenceScore, 1e-9)
}

func TestSummarizeReportsUsageAndCriteriaAgainstReference(t *testing.T) {
	attempts := []RunAttempt{
		{Scenario: "a", AgentRan: true, Turns: 4, Prompt: 1000, Completion: 100, PeakContext: 600, Criteria: []Check{{ID: "ready", Passed: true}, {ID: "svc", Passed: false}}},
		{Scenario: "a", AgentRan: true, Turns: 6, Prompt: 3000, Completion: 300, PeakContext: 1000, Overflow: true, Criteria: []Check{{ID: "ready", Passed: true}, {ID: "svc", Passed: true}}},
		{Scenario: "a", Criteria: []Check{{ID: "ready", Passed: false}}},
	}
	reference := []RunAttempt{
		{Scenario: "a", AgentRan: true, Turns: 10, Prompt: 500, Completion: 50, Criteria: []Check{{ID: "ready", Passed: false}, {ID: "svc", Passed: false}}},
		{Scenario: "z", AgentRan: true, Turns: 99, Criteria: []Check{{ID: "other", Passed: true}}},
	}

	summary := Summarize("rag", attempts, reference)

	assert.Equal(t, Usage{Attempts: 2, Turns: 5, Prompt: 2000, Completion: 200, PeakContext: 800, MaxPeakContext: 1000, Overflows: 1}, summary.Usage)
	require.NotNil(t, summary.ReferenceUsage)
	assert.Equal(t, Usage{Attempts: 1, Turns: 10, Prompt: 500, Completion: 50}, *summary.ReferenceUsage)
	assert.Equal(t, []CriterionRate{
		{Scenario: "a", ID: "ready", Passed: 2, Total: 3, ReferencePassed: 0, ReferenceTotal: 1, ReferenceAvailable: true},
		{Scenario: "a", ID: "svc", Passed: 1, Total: 2, ReferencePassed: 0, ReferenceTotal: 1, ReferenceAvailable: true},
	}, summary.Criteria)
}
