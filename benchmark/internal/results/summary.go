package results

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
)

type conditionAccumulator struct {
	expectedAttempts       int
	attemptCount           int
	fullSuccessCount       int
	scoreTotal             float64
	errorCount             int
	validationPassedCount  int
	validationAttemptCount int
	terminations           map[string]int
	skillLoads             map[string]int
	referenceLoadCount     int
	scenarios              map[string]*scenarioAccumulator
}

type scenarioAccumulator struct {
	attemptCount     int
	fullSuccessCount int
	scoreTotal       float64
}

func expectedAttemptCount(metadata RunMetadata) int {
	if metadata.ExpectedAttempts > 0 {
		return metadata.ExpectedAttempts
	}
	conditionCount := len(metadata.Agents)
	if conditionCount == 0 {
		conditionCount = 1
	}
	return len(metadata.Scenarios) * metadata.RepeatCount * conditionCount
}

func runType(metadata RunMetadata) string {
	if metadata.RunType != "" {
		return metadata.RunType
	}
	if len(metadata.Agents) > 0 {
		return "benchmark"
	}
	return "validation"
}

func summarizeRun(runDir string, metadata RunMetadata) (RunSummary, error) {
	metadata.ExpectedAttempts = expectedAttemptCount(metadata)
	summary := RunSummary{
		ExpectedAttempts: metadata.ExpectedAttempts,
		ByCondition:      make(map[string]ConditionSummary),
	}
	accumulators := make(map[string]*conditionAccumulator)

	if err := filepath.WalkDir(runDir, func(path string, entry os.DirEntry, walkErr error) error {
		if walkErr != nil {
			return walkErr
		}
		if entry.IsDir() || entry.Name() == "run.json" || filepath.Ext(entry.Name()) != ".json" {
			return nil
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return fmt.Errorf("read attempt result %s: %w", path, err)
		}
		var header struct {
			CaseID string `json:"case_id"`
		}
		if err := json.Unmarshal(data, &header); err != nil {
			return fmt.Errorf("decode attempt result %s: %w", path, err)
		}
		if header.CaseID != "" {
			var result ValidationAttemptResult
			if err := json.Unmarshal(data, &result); err != nil {
				return fmt.Errorf("decode validation result %s: %w", path, err)
			}
			condition := result.Condition
			if condition == "" {
				condition = "validation"
			}
			accumulator := conditionFor(accumulators, condition, metadata)
			accumulator.add(result.ScenarioID, result.Grading.FullSuccess, result.Grading.Score, result.Error, "", nil)
			accumulator.validationAttemptCount++
			if result.Passed {
				accumulator.validationPassedCount++
			}
			summary.AttemptsRecorded++
			if result.Error != "" {
				summary.ErrorCount++
			}
			return nil
		}

		var result AttemptResult
		if err := json.Unmarshal(data, &result); err != nil {
			return fmt.Errorf("decode attempt result %s: %w", path, err)
		}
		if result.ScenarioID == "" {
			return nil
		}
		condition := result.Condition
		if condition == "" {
			condition = "unknown"
		}
		termination := ""
		if result.Agent != nil {
			termination = result.Agent.Termination
		}
		accumulator := conditionFor(accumulators, condition, metadata)
		accumulator.add(result.ScenarioID, result.Grading.FullSuccess, result.Grading.Score, result.Error, termination, result.Agent)
		summary.AttemptsRecorded++
		if result.Error != "" {
			summary.ErrorCount++
		}
		return nil
	}); err != nil {
		return RunSummary{}, fmt.Errorf("summarize run %s: %w", metadata.RunID, err)
	}

	for condition, accumulator := range accumulators {
		conditionSummary := accumulator.summary(runType(metadata), metadata)
		summary.ByCondition[condition] = conditionSummary
	}
	return summary, nil
}

func conditionFor(accumulators map[string]*conditionAccumulator, condition string, metadata RunMetadata) *conditionAccumulator {
	if accumulator := accumulators[condition]; accumulator != nil {
		return accumulator
	}
	expected := metadata.RepeatCount * len(metadata.Scenarios)
	if runType(metadata) == "validation" {
		expected = expectedAttemptCount(metadata)
	}
	accumulator := &conditionAccumulator{
		expectedAttempts: expected,
		terminations:     make(map[string]int),
		skillLoads:       make(map[string]int),
		scenarios:        make(map[string]*scenarioAccumulator),
	}
	accumulators[condition] = accumulator
	return accumulator
}

func (a *conditionAccumulator) add(scenarioID string, fullSuccess bool, score float64, errorText, termination string, agent *common.Result) {
	a.attemptCount++
	a.scoreTotal += score
	if fullSuccess {
		a.fullSuccessCount++
	}
	if errorText != "" {
		a.errorCount++
	}
	if termination != "" {
		a.terminations[termination]++
	}
	if agent != nil {
		for _, toolCall := range agent.ToolCalls {
			switch toolCall.Call.Name {
			case "load_skill":
				if toolCall.Error != "" {
					continue
				}
				var arguments struct {
					Name string `json:"name"`
				}
				if json.Unmarshal([]byte(toolCall.Call.Arguments), &arguments) != nil || strings.TrimSpace(arguments.Name) == "" {
					continue
				}
				a.skillLoads[arguments.Name]++
			case "load_reference":
				if toolCall.Error == "" {
					a.referenceLoadCount++
				}
			}
		}
	}
	scenario := a.scenarios[scenarioID]
	if scenario == nil {
		scenario = &scenarioAccumulator{}
		a.scenarios[scenarioID] = scenario
	}
	scenario.attemptCount++
	scenario.scoreTotal += score
	if fullSuccess {
		scenario.fullSuccessCount++
	}
}

func (a *conditionAccumulator) summary(kind string, metadata RunMetadata) ConditionSummary {
	result := ConditionSummary{
		ExpectedAttempts:      a.expectedAttempts,
		AttemptCount:          a.attemptCount,
		FullSuccessCount:      a.fullSuccessCount,
		ErrorCount:            a.errorCount,
		ValidationPassedCount: a.validationPassedCount,
		TerminationCounts:     a.terminations,
		SkillLoadCounts:       a.skillLoads,
		ReferenceLoadCount:    a.referenceLoadCount,
		Scenarios:             make(map[string]ScenarioSummary, len(a.scenarios)),
	}
	if a.attemptCount > 0 {
		result.FullSuccessRate = float64(a.fullSuccessCount) / float64(a.attemptCount)
		result.MeanScore = a.scoreTotal / float64(a.attemptCount)
	}
	if a.validationAttemptCount > 0 {
		rate := float64(a.validationPassedCount) / float64(a.validationAttemptCount)
		result.ValidationPassRate = &rate
	}
	for scenarioID, accumulator := range a.scenarios {
		expected := 0
		if kind == "benchmark" {
			expected = metadata.RepeatCount
		}
		scenario := ScenarioSummary{
			ExpectedAttempts: expected,
			AttemptCount:     accumulator.attemptCount,
			FullSuccessCount: accumulator.fullSuccessCount,
		}
		if accumulator.attemptCount > 0 {
			scenario.FullSuccessRate = float64(accumulator.fullSuccessCount) / float64(accumulator.attemptCount)
			scenario.MeanScore = accumulator.scoreTotal / float64(accumulator.attemptCount)
		}
		result.Scenarios[scenarioID] = scenario
	}
	if kind == "benchmark" && a.expectedAttempts > 0 && a.attemptCount == a.expectedAttempts && len(a.scenarios) == len(metadata.Scenarios) {
		var macroScore float64
		complete := true
		for _, scenarioID := range metadata.Scenarios {
			scenario := a.scenarios[scenarioID]
			if scenario == nil || scenario.attemptCount != metadata.RepeatCount {
				complete = false
				break
			}
			macroScore += scenario.scoreTotal / float64(scenario.attemptCount)
		}
		if complete && len(metadata.Scenarios) > 0 {
			macroScore /= float64(len(metadata.Scenarios))
			result.MacroAverageScore = &macroScore
		}
	}
	return result
}

func refreshStudyResults(root string) error {
	entries, err := os.ReadDir(root)
	if err != nil {
		return fmt.Errorf("read results directory: %w", err)
	}
	runs := make([]StudyRun, 0, len(entries))
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		runDir := filepath.Join(root, entry.Name())
		metadataPath := filepath.Join(runDir, "run.json")
		data, err := os.ReadFile(metadataPath)
		if errors.Is(err, os.ErrNotExist) {
			continue
		}
		if err != nil {
			return fmt.Errorf("read run metadata %s: %w", metadataPath, err)
		}
		var metadata RunMetadata
		if err := json.Unmarshal(data, &metadata); err != nil {
			return fmt.Errorf("decode run metadata %s: %w", metadataPath, err)
		}
		if metadata.RunType == "" {
			metadata.RunType = runType(metadata)
		}
		if metadata.ExpectedAttempts == 0 {
			metadata.ExpectedAttempts = expectedAttemptCount(metadata)
		}
		if metadata.State == "" {
			if metadata.CompletedAt == nil {
				metadata.State = RunStateRunning
			} else {
				metadata.State = RunStateCompleted
			}
		}
		if metadata.Summary == nil {
			summary, err := summarizeRun(runDir, metadata)
			if err != nil {
				return err
			}
			metadata.Summary = &summary
		}
		runs = append(runs, StudyRun{
			RunID:              metadata.RunID,
			RunType:            metadata.RunType,
			State:              metadata.State,
			StartedAt:          metadata.StartedAt,
			CompletedAt:        metadata.CompletedAt,
			RepositoryRevision: metadata.RepositoryRevision,
			WorkingTreeDirty:   metadata.WorkingTreeDirty,
			ExpectedAttempts:   metadata.ExpectedAttempts,
			Agents:             metadata.Agents,
			Scenarios:          metadata.Scenarios,
			RepeatCount:        metadata.RepeatCount,
			Summary:            metadata.Summary,
			Path:               filepath.ToSlash(filepath.Join(entry.Name(), "run.json")),
		})
	}
	slices.SortFunc(runs, func(a, b StudyRun) int {
		if a.StartedAt.Before(b.StartedAt) {
			return -1
		}
		if a.StartedAt.After(b.StartedAt) {
			return 1
		}
		return strings.Compare(a.RunID, b.RunID)
	})
	return writeJSON(filepath.Join(root, "results.json"), StudyResults{
		UpdatedAt: time.Now().UTC(),
		Runs:      runs,
	})
}
