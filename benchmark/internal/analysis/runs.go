// Package analysis scores finished benchmark runs evaluator-side: RAG search
// use and retrieval against scenario sources, cost and per-criterion outcomes.
package analysis

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

// SearchToolName is the RAG condition's search tool, as recorded in traces.
const SearchToolName = "search_docs"

// Search is one recorded search_docs call. SourceRank is the rank of the first
// hit from the scenario's source file, or 0 when the source was not returned.
type Search struct {
	Call         int      `json:"call"`
	Query        string   `json:"query"`
	BeforeChange bool     `json:"before_change"`
	SourceRank   int      `json:"source_rank"`
	Truncated    bool     `json:"truncated,omitempty"`
	Paths        []string `json:"paths"`
	Error        string   `json:"error,omitempty"`
}

// Check is the outcome of one grading criterion.
type Check struct {
	ID     string `json:"id"`
	Passed bool   `json:"passed"`
}

// RunAttempt joins one benchmark attempt with its scenario's evaluator-only
// source reference. FirstChange is the 0-based tool-call index of the first
// Bash command that changes the cluster, or -1. Changes counts such commands
// and FailedChanges those that exited non-zero; Edits counts `kubectl edit`,
// which cannot work without an editor, and RolloutStatus records whether the
// agent ran `kubectl rollout status`.
type RunAttempt struct {
	Run           string   `json:"run"`
	Condition     string   `json:"condition"`
	Scenario      string   `json:"scenario"`
	Attempt       int      `json:"attempt"`
	Score         float64  `json:"score"`
	FullSuccess   bool     `json:"full_success"`
	Termination   string   `json:"termination,omitempty"`
	Error         bool     `json:"error,omitempty"`
	AgentRan      bool     `json:"agent_ran"`
	ToolCalls     int      `json:"tool_calls"`
	Turns         int      `json:"turns"`
	Prompt        int      `json:"prompt_tokens"`
	Completion    int      `json:"completion_tokens"`
	PeakContext   int      `json:"peak_context_tokens"`
	Overflow      bool     `json:"context_overflow,omitempty"`
	Criteria      []Check  `json:"criteria"`
	FirstChange   int      `json:"first_change"`
	Changes       int      `json:"changes"`
	FailedChanges int      `json:"failed_changes"`
	Edits         int      `json:"edits"`
	RolloutStatus bool     `json:"rollout_status"`
	Source        string   `json:"source"`
	Searches      []Search `json:"searches,omitempty"`
}

func (a RunAttempt) SourceRetrieved() bool {
	return slices.ContainsFunc(a.Searches, func(search Search) bool { return search.SourceRank > 0 })
}

var sourcePath = regexp.MustCompile("Source path: `([^`]+)`")

// LoadSources reads the evaluator-only source path of every scenario under
// root, keyed by scenario directory name (the scenario ID).
func LoadSources(root string) (map[string]string, error) {
	files, err := filepath.Glob(filepath.Join(root, "*", "*", "source.md"))
	if err != nil {
		return nil, err
	}
	sources := make(map[string]string, len(files))
	for _, file := range files {
		data, err := os.ReadFile(file)
		if err != nil {
			return nil, err
		}
		match := sourcePath.FindSubmatch(data)
		if match == nil {
			return nil, fmt.Errorf("%s has no source path", file)
		}
		sources[filepath.Base(filepath.Dir(file))] = string(match[1])
	}
	if len(sources) == 0 {
		return nil, fmt.Errorf("no source.md files under %s", root)
	}
	return sources, nil
}

// LoadSubset reads scenario paths such as scenarios/area/id, one per line,
// and accepts their IDs; an empty path accepts every scenario.
func LoadSubset(path string) (func(string) bool, error) {
	if path == "" {
		return func(string) bool { return true }, nil
	}
	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	ids := map[string]bool{}
	for _, line := range strings.Fields(string(data)) {
		ids[filepath.Base(line)] = true
	}
	return func(id string) bool { return ids[id] }, nil
}

// LoadRunAttempts reads every benchmark attempt of a run.
func LoadRunAttempts(root, runID string, sources map[string]string) ([]RunAttempt, error) {
	run, err := results.LoadRun(root, runID)
	if err != nil {
		return nil, err
	}
	attempts := make([]RunAttempt, 0, len(run.Attempts))
	for _, ref := range run.Attempts {
		loaded, err := results.LoadAttempt(ref, run.Metadata.RunType)
		if err != nil {
			return nil, err
		}
		if loaded.Benchmark == nil {
			return nil, fmt.Errorf("run %s is not a benchmark run", runID)
		}
		attempt, err := newRunAttempt(runID, *loaded.Benchmark, sources)
		if err != nil {
			return nil, err
		}
		attempts = append(attempts, attempt)
	}
	return attempts, nil
}

func newRunAttempt(runID string, result results.AttemptResult, sources map[string]string) (RunAttempt, error) {
	source, ok := sources[result.ScenarioID]
	if !ok {
		return RunAttempt{}, fmt.Errorf("scenario %s has no source reference", result.ScenarioID)
	}
	attempt := RunAttempt{
		Run:         runID,
		Condition:   result.Condition,
		Scenario:    result.ScenarioID,
		Attempt:     result.Attempt,
		Score:       result.Grading.Score,
		FullSuccess: result.Grading.FullSuccess,
		Error:       result.Error != "",
		FirstChange: -1,
		Source:      source,
	}
	for _, criterion := range result.Grading.Criteria {
		attempt.Criteria = append(attempt.Criteria, Check{ID: criterion.ID, Passed: criterion.Passed})
	}
	if result.Agent == nil {
		return attempt, nil
	}
	attempt.AgentRan = true
	attempt.Termination = result.Agent.Termination
	attempt.ToolCalls = len(result.Agent.ToolCalls)
	attempt.Turns = result.Agent.Turns
	attempt.Prompt = result.Agent.TokenUsage.PromptTokens
	attempt.Completion = result.Agent.TokenUsage.CompletionTokens
	attempt.PeakContext = result.Agent.TokenUsage.PeakContextTokens
	attempt.Overflow = result.Agent.ContextOverflow
	for index, call := range result.Agent.ToolCalls {
		switch call.Call.Name {
		case "bash":
			var arguments struct {
				Command string `json:"command"`
			}
			if json.Unmarshal([]byte(call.Call.Arguments), &arguments) != nil {
				continue
			}
			var output struct {
				ExitCode int `json:"exit_code"`
			}
			if err := decodeDetails(call.Details, &output); err != nil {
				return RunAttempt{}, fmt.Errorf("%s attempt %d: %w", result.ScenarioID, result.Attempt, err)
			}
			attempt.observeCommand(index, arguments.Command, output.ExitCode)
		case SearchToolName:
			search, err := newSearch(index, call.Details, source)
			if err != nil {
				return RunAttempt{}, fmt.Errorf("%s attempt %d: %w", result.ScenarioID, result.Attempt, err)
			}
			search.BeforeChange = attempt.FirstChange < 0
			search.Error = call.Error
			attempt.Searches = append(attempt.Searches, search)
		}
	}
	return attempt, nil
}

func (a *RunAttempt) observeCommand(index int, command string, exitCode int) {
	for _, args := range kubectlCommands(command) {
		switch {
		case writes(args):
			if a.FirstChange < 0 {
				a.FirstChange = index
			}
			a.Changes++
			if exitCode != 0 {
				a.FailedChanges++
			}
			if args[0] == "edit" {
				a.Edits++
			}
		case len(args) > 1 && args[0] == "rollout" && args[1] == "status":
			a.RolloutStatus = true
		}
	}
}

// decodeDetails decodes tool call details, a JSON object once loaded from an
// attempt file; nil details leave target unchanged.
func decodeDetails(details, target any) error {
	if details == nil {
		return nil
	}
	encoded, err := json.Marshal(details)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(encoded, target); err != nil {
		return fmt.Errorf("decode tool call details: %w", err)
	}
	return nil
}

// newSearch decodes the rag SearchEvidence recorded as tool call details.
func newSearch(call int, details any, source string) (Search, error) {
	search := Search{Call: call}
	var evidence struct {
		Query     string `json:"query"`
		Truncated bool   `json:"truncated"`
		Hits      []struct {
			Rank  int `json:"rank"`
			Chunk struct {
				Path string `json:"path"`
			} `json:"chunk"`
		} `json:"hits"`
	}
	if err := decodeDetails(details, &evidence); err != nil {
		return Search{}, err
	}
	search.Query = evidence.Query
	search.Truncated = evidence.Truncated
	for _, hit := range evidence.Hits {
		search.Paths = append(search.Paths, hit.Chunk.Path)
		if search.SourceRank == 0 && hit.Chunk.Path == source {
			search.SourceRank = hit.Rank
		}
	}
	return search, nil
}

// changeVerbs are kubectl subcommands that write to the cluster.
var changeVerbs = map[string]bool{
	"apply": true, "patch": true, "edit": true, "set": true, "delete": true, "create": true,
	"replace": true, "scale": true, "label": true, "annotate": true, "taint": true,
	"cordon": true, "uncordon": true, "drain": true, "expose": true, "autoscale": true,
}

// flagsWithValue are kubectl global flags whose value is a separate word.
var flagsWithValue = map[string]bool{"-n": true, "--namespace": true, "--context": true, "--kubeconfig": true}

var commandSeparator = regexp.MustCompile(`\|\||&&|[|;&\n]`)

// ChangesCluster reports whether a Bash command runs a kubectl subcommand that
// writes to the cluster. It is a heuristic over the command text, used only to
// order searches against the first repair attempt and to count changes.
func ChangesCluster(command string) bool {
	return slices.ContainsFunc(kubectlCommands(command), writes)
}

func writes(args []string) bool {
	return changeVerbs[args[0]] || args[0] == "rollout" && len(args) > 1 && (args[1] == "restart" || args[1] == "undo")
}

// kubectlCommands returns the positional arguments of every kubectl
// invocation in a shell command, skipping --dry-run invocations.
func kubectlCommands(command string) [][]string {
	var commands [][]string
	for _, segment := range commandSeparator.Split(command, -1) {
		if strings.Contains(segment, "--dry-run") {
			continue
		}
		words := strings.Fields(segment)
		start := slices.Index(words, "kubectl")
		if start < 0 {
			continue
		}
		var args []string
		for index := start + 1; index < len(words); index++ {
			word := words[index]
			if strings.HasPrefix(word, "-") {
				if flagsWithValue[word] {
					index++
				}
				continue
			}
			args = append(args, word)
		}
		if len(args) > 0 {
			commands = append(commands, args)
		}
	}
	return commands
}
