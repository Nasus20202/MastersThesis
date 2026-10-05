package analysis

import (
	"cmp"
	"maps"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/jsonfile"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval/evaluation"
)

// RunSummary describes one condition of one run. Macro scores are means of
// per-scenario mean scores. Reference columns are the reference condition's
// results on the same scenarios.
type RunSummary struct {
	Label                string          `json:"label"`
	Attempts             int             `json:"attempts"`
	Macro                float64         `json:"macro"`
	FullSuccess          int             `json:"full_success"`
	AgentNotRun          int             `json:"agent_not_run"`
	SearchAttempts       int             `json:"search_attempts"`
	Searches             int             `json:"searches"`
	SearchesBeforeChange int             `json:"searches_before_change"`
	SearchErrors         int             `json:"search_errors"`
	SourceHits           int             `json:"source_hits"`
	SourceAttempts       int             `json:"source_retrieved_attempts"`
	MacroDifference      *Difference     `json:"macro_difference,omitempty"`
	Usage                Usage           `json:"usage"`
	ReferenceUsage       *Usage          `json:"reference_usage,omitempty"`
	Criteria             []CriterionRate `json:"criteria"`
	Groups               []OutcomeGroup  `json:"groups"`
	Scenarios            []ScenarioUsage `json:"scenarios"`
}

// Difference is the macro score minus the reference's on the scenarios both
// cover, with a 95% paired bootstrap interval over scenarios.
type Difference struct {
	Scenarios int     `json:"scenarios"`
	Mean      float64 `json:"mean"`
	Low       float64 `json:"ci_low"`
	High      float64 `json:"ci_high"`
}

// Usage is the mean model cost of an attempt and how the agent changed and
// checked the cluster, over attempts where the agent ran. Unconfirmed counts
// attempts the agent completed without full success: it reported a repair the
// grader did not confirm. PeakContext and MaxPeakContext are the mean and
// largest per-attempt peak context; Overflows counts attempts that ran out of
// context.
type Usage struct {
	Attempts       int     `json:"attempts"`
	Turns          float64 `json:"turns"`
	Prompt         float64 `json:"prompt_tokens"`
	Completion     float64 `json:"completion_tokens"`
	PeakContext    float64 `json:"peak_context_tokens"`
	MaxPeakContext int     `json:"max_peak_context_tokens"`
	Overflows      int     `json:"context_overflows"`
	Changes        int     `json:"changes"`
	FailedChanges  int     `json:"failed_changes"`
	Edits          int     `json:"edits"`
	RolloutStatus  int     `json:"rollout_status_attempts"`
	Unconfirmed    int     `json:"unconfirmed_attempts"`
}

// CriterionRate is the pass rate of one scenario criterion. Criterion IDs are
// scenario-local, so rates are never merged across scenarios.
type CriterionRate struct {
	Scenario           string `json:"scenario"`
	ID                 string `json:"id"`
	Passed             int    `json:"passed"`
	Total              int    `json:"total"`
	ReferencePassed    int    `json:"reference_passed"`
	ReferenceTotal     int    `json:"reference_total"`
	ReferenceAvailable bool   `json:"reference_available"`
}

type OutcomeGroup struct {
	Name               string  `json:"name"`
	Attempts           int     `json:"attempts"`
	Macro              float64 `json:"macro"`
	FullSuccess        int     `json:"full_success"`
	Scenarios          int     `json:"scenarios"`
	ReferenceMacro     float64 `json:"reference_macro"`
	ReferenceFull      int     `json:"reference_full_success"`
	ReferenceAttempts  int     `json:"reference_attempts"`
	ReferenceAvailable bool    `json:"reference_available"`
}

type ScenarioUsage struct {
	Scenario        string   `json:"scenario"`
	Sources         []string `json:"sources"`
	Attempts        int      `json:"attempts"`
	MeanScore       float64  `json:"mean_score"`
	FullSuccess     int      `json:"full_success"`
	SearchAttempts  int      `json:"search_attempts"`
	Searches        int      `json:"searches"`
	SourceRetrieved int      `json:"source_retrieved_attempts"`
	ReferenceScore  *float64 `json:"reference_mean_score,omitempty"`
}

// Outcome groups split attempts by what search returned to them.
const (
	GroupNoSearch = "no search"
	GroupMissed   = "searched, source not retrieved"
	GroupFound    = "source retrieved"
)

// Summarize aggregates attempts of one condition against optional reference
// attempts (another condition, usually the prompt agent).
func Summarize(label string, attempts, reference []RunAttempt) RunSummary {
	summary := RunSummary{Label: label, Attempts: len(attempts)}
	summary.Macro, summary.FullSuccess, _ = outcome(attempts)
	for _, attempt := range attempts {
		if !attempt.AgentRan {
			summary.AgentNotRun++
		}
		if len(attempt.Searches) > 0 {
			summary.SearchAttempts++
		}
		if attempt.SourceRetrieved() {
			summary.SourceAttempts++
		}
		for _, search := range attempt.Searches {
			summary.Searches++
			if search.BeforeChange {
				summary.SearchesBeforeChange++
			}
			if search.Error != "" {
				summary.SearchErrors++
			}
			if search.SourceRank > 0 {
				summary.SourceHits++
			}
		}
	}
	groups := map[string][]RunAttempt{}
	for _, attempt := range attempts {
		name := GroupNoSearch
		if attempt.SourceRetrieved() {
			name = GroupFound
		} else if len(attempt.Searches) > 0 {
			name = GroupMissed
		}
		groups[name] = append(groups[name], attempt)
	}
	for _, name := range []string{GroupNoSearch, GroupMissed, GroupFound} {
		members := groups[name]
		group := OutcomeGroup{Name: name, Attempts: len(members)}
		group.Macro, group.FullSuccess, group.Scenarios = outcome(members)
		if matched := onScenarios(reference, scenarioSet(members)); len(matched) > 0 {
			group.ReferenceAvailable = true
			group.ReferenceAttempts = len(matched)
			group.ReferenceMacro, group.ReferenceFull, _ = outcome(matched)
		}
		summary.Groups = append(summary.Groups, group)
	}
	summary.Scenarios = scenarioUsage(attempts, reference)
	summary.Usage = usage(attempts)
	matched := onScenarios(reference, scenarioSet(attempts))
	if len(matched) > 0 {
		referenceUsage := usage(matched)
		summary.ReferenceUsage = &referenceUsage
	}
	summary.Criteria = criterionRates(attempts, matched)
	summary.MacroDifference = macroDifference(attempts, reference)
	return summary
}

func macroDifference(attempts, reference []RunAttempt) *Difference {
	means, referenceMeans := scenarioMeans(attempts), scenarioMeans(reference)
	var differences []float64
	for scenario, score := range means {
		if referenceScore, ok := referenceMeans[scenario]; ok {
			differences = append(differences, score-referenceScore)
		}
	}
	return pairedDifference(differences)
}

// pairedDifference is the mean of per-scenario differences with its bootstrap
// interval, or nil without differences.
func pairedDifference(differences []float64) *Difference {
	if len(differences) == 0 {
		return nil
	}
	// Map iteration order is random; the seeded resampling needs a fixed order.
	slices.Sort(differences)
	result := &Difference{Scenarios: len(differences)}
	for _, difference := range differences {
		result.Mean += difference
	}
	result.Mean /= float64(len(differences))
	result.Low, result.High = evaluation.BootstrapInterval(differences)
	return result
}

func scenarioUsage(attempts, reference []RunAttempt) []ScenarioUsage {
	byScenario := map[string]*ScenarioUsage{}
	var order []string
	for _, attempt := range attempts {
		usage, ok := byScenario[attempt.Scenario]
		if !ok {
			usage = &ScenarioUsage{Scenario: attempt.Scenario, Sources: attempt.Sources}
			byScenario[attempt.Scenario] = usage
			order = append(order, attempt.Scenario)
		}
		usage.Attempts++
		usage.MeanScore += attempt.Score
		if attempt.FullSuccess {
			usage.FullSuccess++
		}
		if len(attempt.Searches) > 0 {
			usage.SearchAttempts++
		}
		usage.Searches += len(attempt.Searches)
		if attempt.SourceRetrieved() {
			usage.SourceRetrieved++
		}
	}
	referenceScores := scenarioMeans(reference)
	slices.Sort(order)
	rows := make([]ScenarioUsage, 0, len(order))
	for _, scenario := range order {
		usage := byScenario[scenario]
		usage.MeanScore /= float64(usage.Attempts)
		if score, ok := referenceScores[scenario]; ok {
			usage.ReferenceScore = &score
		}
		rows = append(rows, *usage)
	}
	return rows
}

func usage(attempts []RunAttempt) Usage {
	var result Usage
	for _, attempt := range attempts {
		if !attempt.AgentRan {
			continue
		}
		result.Attempts++
		result.Turns += float64(attempt.Turns)
		result.Prompt += float64(attempt.Prompt)
		result.Completion += float64(attempt.Completion)
		result.PeakContext += float64(attempt.PeakContext)
		result.MaxPeakContext = max(result.MaxPeakContext, attempt.PeakContext)
		if attempt.Overflow {
			result.Overflows++
		}
		result.Changes += attempt.Changes
		result.FailedChanges += attempt.FailedChanges
		result.Edits += attempt.Edits
		if attempt.RolloutStatus {
			result.RolloutStatus++
		}
		if attempt.Termination == "completed" && !attempt.FullSuccess {
			result.Unconfirmed++
		}
	}
	if result.Attempts > 0 {
		count := float64(result.Attempts)
		result.Turns /= count
		result.Prompt /= count
		result.Completion /= count
		result.PeakContext /= count
	}
	return result
}

type criterionKey struct{ scenario, id string }

type passCount struct{ passed, total int }

func countCriteria(attempts []RunAttempt) map[criterionKey]*passCount {
	counts := map[criterionKey]*passCount{}
	for _, attempt := range attempts {
		for _, check := range attempt.Criteria {
			key := criterionKey{attempt.Scenario, check.ID}
			if counts[key] == nil {
				counts[key] = &passCount{}
			}
			counts[key].total++
			if check.Passed {
				counts[key].passed++
			}
		}
	}
	return counts
}

func criterionRates(attempts, reference []RunAttempt) []CriterionRate {
	counts, referenceCounts := countCriteria(attempts), countCriteria(reference)
	rates := make([]CriterionRate, 0, len(counts))
	for key, count := range counts {
		rate := CriterionRate{Scenario: key.scenario, ID: key.id, Passed: count.passed, Total: count.total}
		if other, ok := referenceCounts[key]; ok {
			rate.ReferenceAvailable = true
			rate.ReferencePassed, rate.ReferenceTotal = other.passed, other.total
		}
		rates = append(rates, rate)
	}
	slices.SortFunc(rates, func(a, b CriterionRate) int {
		return cmp.Or(strings.Compare(a.Scenario, b.Scenario), strings.Compare(a.ID, b.ID))
	})
	return rates
}

// Macro is the mean of per-scenario mean scores.
func Macro(attempts []RunAttempt) float64 {
	macro, _, _ := outcome(attempts)
	return macro
}

// outcome returns the macro score, full-success count and scenario count.
func outcome(attempts []RunAttempt) (float64, int, int) {
	means := scenarioMeans(attempts)
	full := 0
	for _, attempt := range attempts {
		if attempt.FullSuccess {
			full++
		}
	}
	if len(means) == 0 {
		return 0, full, 0
	}
	// Summing in scenario order keeps the result identical between runs.
	total := 0.0
	for _, scenario := range slices.Sorted(maps.Keys(means)) {
		total += means[scenario]
	}
	return total / float64(len(means)), full, len(means)
}

func scenarioMeans(attempts []RunAttempt) map[string]float64 {
	sums := map[string]float64{}
	counts := map[string]int{}
	for _, attempt := range attempts {
		sums[attempt.Scenario] += attempt.Score
		counts[attempt.Scenario]++
	}
	for scenario, count := range counts {
		sums[scenario] /= float64(count)
	}
	return sums
}

func scenarioSet(attempts []RunAttempt) map[string]bool {
	set := map[string]bool{}
	for _, attempt := range attempts {
		set[attempt.Scenario] = true
	}
	return set
}

func onScenarios(attempts []RunAttempt, scenarios map[string]bool) []RunAttempt {
	var matched []RunAttempt
	for _, attempt := range attempts {
		if scenarios[attempt.Scenario] {
			matched = append(matched, attempt)
		}
	}
	return matched
}

// WriteRunAnalysis stores summary.json and attempts.jsonl with one attempt per
// line.
func WriteRunAnalysis(dir string, summaries []RunSummary, attempts []RunAttempt) error {
	if err := os.MkdirAll(dir, 0o750); err != nil {
		return err
	}
	if err := jsonfile.Write(filepath.Join(dir, "summary.json"), summaries); err != nil {
		return err
	}
	return jsonfile.WriteLines(filepath.Join(dir, "attempts.jsonl"), attempts)
}
