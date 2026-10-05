package analysis

import (
	"encoding/json"
	"os"
	"path/filepath"
	"slices"
	"testing"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/orchestration"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestKubectlWrites(t *testing.T) {
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
		assert.Equal(t, want, slices.ContainsFunc(kubectlCommands(command), writes), command)
	}
}

func TestLoadSources(t *testing.T) {
	root := t.TempDir()
	writeScenario(t, filepath.Join(root, "networking", "service-port"), "service-port", "sources:\n  - path: docs/service.md\n  - path: docs/dns.md\n")
	writeScenario(t, filepath.Join(root, "networking", "no-source"), "no-source", "")

	sources, err := LoadSources(root)

	require.NoError(t, err)
	assert.Equal(t, map[string][]string{"service-port": {"docs/service.md", "docs/dns.md"}}, sources)
}

func writeScenario(t *testing.T, dir, id, sources string) {
	t.Helper()
	definition := "id: " + id + "\ntitle: Test\ntask: Fix it.\n" + sources +
		"prepare: [{program: prepare}]\nverify_clean: [{program: check}]\ngrading: [{id: ready, weight: 1, check: {program: check}}]\n"
	require.NoError(t, os.MkdirAll(dir, 0o750))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "scenario.yaml"), []byte(definition), 0o600))
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

	attempt, err := NewRunAttempt("run", results.Attempt{Benchmark: &result}, []string{"intro.md", "service.md"})

	require.NoError(t, err)
	assert.True(t, attempt.AgentRan)
	assert.Equal(t, 2, attempt.FirstChange)
	require.Len(t, attempt.Searches, 2)
	assert.Equal(t, Search{Call: 1, Query: "service port", BeforeChange: true, SourceRank: 2, Paths: []string{"other.md", "service.md"}}, attempt.Searches[0])
	assert.False(t, attempt.Searches[1].BeforeChange)
	assert.Zero(t, attempt.Searches[1].SourceRank)
	assert.Equal(t, "search failed", attempt.Searches[1].Error)
	assert.True(t, attempt.SourceRetrieved())
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

	attempt, err := NewRunAttempt("run", results.Attempt{Benchmark: &result}, []string{"app.md"})
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

func TestNewRunAttemptRecordsCostAndDuration(t *testing.T) {
	timings := func(predicted int, ms float64, drafted, accepted int) common.ResponseEvidence {
		return common.ResponseEvidence{Response: inference.Result{Timings: &inference.Timings{PredictedN: predicted, PredictedMS: ms, DraftN: drafted, DraftNAccepted: accepted}}}
	}
	grading := orchestration.GradingResult{Score: 1, Criteria: []orchestration.CriterionResult{{ID: "ready", Passed: true, DurationSeconds: 2}, {ID: "svc", DurationSeconds: 3}}}
	benchmark := results.AttemptResult{ScenarioID: "app", Grading: grading, Agent: &common.Result{
		DurationSeconds: 40,
		TokenUsage:      common.TokenUsage{TotalTokens: 1200, CachedTokens: 800},
		Responses:       []common.ResponseEvidence{timings(100, 2000, 60, 30), {}, timings(50, 500, 0, 0)},
	}}

	attempt, err := NewRunAttempt("run", results.Attempt{Benchmark: &benchmark}, nil)

	require.NoError(t, err)
	assert.InDelta(t, 40, attempt.Duration, 1e-9)
	assert.Equal(t, 1200, attempt.Tokens)
	assert.Equal(t, 800, attempt.Cached)
	assert.Equal(t, 150, attempt.Predicted)
	assert.InDelta(t, 2.5, attempt.PredictedTime, 1e-9)
	assert.Equal(t, 60, attempt.Drafted)
	assert.Equal(t, 30, attempt.DraftAccepted)

	validation := results.ValidationAttemptResult{Condition: "validation", ScenarioID: "app", Attempt: 2, Grading: grading, Error: "expected score 0"}
	attempt, err = NewRunAttempt("run", results.Attempt{Validation: &validation}, nil)

	require.NoError(t, err)
	assert.Equal(t, RunAttempt{Run: "run", Condition: "validation", Scenario: "app", Attempt: 2, Score: 1, Error: true, Duration: 5, FirstChange: -1,
		Criteria: []Check{{ID: "ready", Passed: true}, {ID: "svc"}}}, attempt)
}

func TestLoadRunAttemptsRequiresScenarioSources(t *testing.T) {
	root := t.TempDir()
	store, err := results.New(root, results.RunMetadata{RunID: "run", Agents: []string{"prompt"}, Parallelism: 1, RepeatCount: 1, Scenarios: []string{"app"}})
	require.NoError(t, err)
	require.NoError(t, store.WriteAttempt(1, "prompt", orchestration.RunResult{ScenarioID: "app", Condition: "prompt"}))

	attempts, err := LoadRunAttempts(root, "run", map[string][]string{"app": {"app.md"}})
	require.NoError(t, err)
	require.Len(t, attempts, 1)
	assert.Equal(t, []string{"app.md"}, attempts[0].Sources)

	_, err = LoadRunAttempts(root, "run", map[string][]string{})
	assert.EqualError(t, err, "scenario app has no sources")
}
