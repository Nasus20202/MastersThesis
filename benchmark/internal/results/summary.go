package results

import "time"

type summaryAccumulator struct {
	metadata         RunMetadata
	attemptsRecorded int
	errorCount       int
	byCondition      map[string]*conditionAccumulator
}

type conditionAccumulator struct {
	expectedAttempts       int
	attemptCount           int
	fullSuccessCount       int
	scoreTotal             float64
	errorCount             int
	validationPassedCount  int
	validationAttemptCount int
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

func newSummaryAccumulator(metadata RunMetadata) *summaryAccumulator {
	metadata.ExpectedAttempts = expectedAttemptCount(metadata)
	accumulator := &summaryAccumulator{
		metadata:    metadata,
		byCondition: make(map[string]*conditionAccumulator),
	}
	conditions := metadata.Agents
	if runType(metadata) == "validation" {
		conditions = []string{"validation"}
	}
	for _, condition := range conditions {
		accumulator.condition(condition)
	}
	return accumulator
}

func (a *summaryAccumulator) condition(name string) *conditionAccumulator {
	if condition := a.byCondition[name]; condition != nil {
		return condition
	}
	expected := a.metadata.RepeatCount * len(a.metadata.Scenarios)
	if runType(a.metadata) == "validation" {
		expected = a.metadata.ExpectedAttempts
	}
	condition := &conditionAccumulator{
		expectedAttempts: expected,
		scenarios:        make(map[string]*scenarioAccumulator, len(a.metadata.Scenarios)),
	}
	for _, scenarioID := range a.metadata.Scenarios {
		condition.scenarios[scenarioID] = &scenarioAccumulator{}
	}
	a.byCondition[name] = condition
	return condition
}

func (a *summaryAccumulator) add(conditionID, scenarioID string, fullSuccess bool, score float64, errorText string, validationPassed *bool) {
	a.attemptsRecorded++
	if errorText != "" {
		a.errorCount++
	}
	condition := a.condition(conditionID)
	condition.attemptCount++
	condition.scoreTotal += score
	if fullSuccess {
		condition.fullSuccessCount++
	}
	if errorText != "" {
		condition.errorCount++
	}
	if validationPassed != nil {
		condition.validationAttemptCount++
		if *validationPassed {
			condition.validationPassedCount++
		}
	}
	scenario := condition.scenarios[scenarioID]
	if scenario == nil {
		scenario = &scenarioAccumulator{}
		condition.scenarios[scenarioID] = scenario
	}
	scenario.attemptCount++
	scenario.scoreTotal += score
	if fullSuccess {
		scenario.fullSuccessCount++
	}
}

func (a *summaryAccumulator) summary(state RunState, completedAt *time.Time) RunSummary {
	result := RunSummary{
		RunID:            a.metadata.RunID,
		State:            state,
		UpdatedAt:        time.Now().UTC(),
		CompletedAt:      completedAt,
		ExpectedAttempts: a.metadata.ExpectedAttempts,
		AttemptsRecorded: a.attemptsRecorded,
		ErrorCount:       a.errorCount,
		ByCondition:      make(map[string]ConditionSummary, len(a.byCondition)),
	}
	for name, accumulator := range a.byCondition {
		result.ByCondition[name] = accumulator.summary(runType(a.metadata), a.metadata)
	}
	return result
}

func (a *conditionAccumulator) summary(kind string, metadata RunMetadata) ConditionSummary {
	result := ConditionSummary{
		ExpectedAttempts:      a.expectedAttempts,
		AttemptCount:          a.attemptCount,
		FullSuccessCount:      a.fullSuccessCount,
		ErrorCount:            a.errorCount,
		ValidationPassedCount: a.validationPassedCount,
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
