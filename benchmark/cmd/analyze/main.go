// Package main is the evaluator-side CLI that scores finished benchmark runs.
package main

import (
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"strings"
	"text/tabwriter"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/analysis"
)

const usage = `Usage:
  analyze --condition RUN_ID/CONDITION ... [--reference RUN_ID/CONDITION ...] [--by-model] [--results DIR] [--scenarios DIR] [--subset FILE] [--out DIR]

With --by-model, all --condition runs are pooled and split by the model that
ran them, each model is compared with the reference attempts of the same
model, and the mean over models is reported.`

func main() {
	if err := run(os.Args[1:], os.Stdout, os.Stderr); err != nil && !errors.Is(err, flag.ErrHelp) {
		fmt.Fprintln(os.Stderr, "analyze:", err)
		os.Exit(1)
	}
}

type stringList []string

func (s *stringList) String() string { return strings.Join(*s, ",") }

func (s *stringList) Set(value string) error {
	*s = append(*s, value)
	return nil
}

func run(args []string, output, logOutput io.Writer) error {
	flags := flag.NewFlagSet("analyze", flag.ContinueOnError)
	flags.SetOutput(logOutput)
	flags.Usage = func() {
		fmt.Fprintln(logOutput, usage)
		fmt.Fprintln(logOutput, "\nOptions:")
		flags.PrintDefaults()
	}
	var conditions stringList
	flags.Var(&conditions, "condition", "RUN_ID/CONDITION to analyze; may be repeated")
	var references stringList
	flags.Var(&references, "reference", "RUN_ID/CONDITION compared on the same scenarios, e.g. the prompt agent; may be repeated and is pooled")
	byModel := flags.Bool("by-model", false, "pool the conditions and report each model and the mean over models")
	resultsRoot := flags.String("results", "results", "benchmark results directory")
	scenarios := flags.String("scenarios", "scenarios", "scenario corpus with evaluator-only source.md files")
	out := flags.String("out", "", "directory for attempts.jsonl and summary.json")
	subset := flags.String("subset", "", "file listing scenario paths, one per line; restricts every condition to those scenarios")
	if err := flags.Parse(args); err != nil {
		return err
	}
	if flags.NArg() > 0 {
		return fmt.Errorf("unexpected arguments: %v", flags.Args())
	}
	if len(conditions) == 0 {
		return errors.New("--condition is required")
	}
	sources, err := analysis.LoadSources(*scenarios)
	if err != nil {
		return err
	}
	include, err := analysis.LoadSubset(*subset)
	if err != nil {
		return err
	}
	runs := map[string][]analysis.RunAttempt{}
	load := func(label string) ([]analysis.RunAttempt, error) {
		runID, condition, ok := strings.Cut(label, "/")
		if !ok {
			return nil, fmt.Errorf("%q is not RUN_ID/CONDITION", label)
		}
		if _, loaded := runs[runID]; !loaded {
			attempts, err := analysis.LoadRunAttempts(*resultsRoot, runID, sources)
			if err != nil {
				return nil, err
			}
			runs[runID] = attempts
		}
		var selected []analysis.RunAttempt
		for _, attempt := range runs[runID] {
			if attempt.Condition == condition && include(attempt.Scenario) {
				selected = append(selected, attempt)
			}
		}
		if len(selected) == 0 {
			return nil, fmt.Errorf("run %s has no %s attempts", runID, condition)
		}
		return selected, nil
	}
	var referenceAttempts []analysis.RunAttempt
	for _, label := range references {
		attempts, err := load(label)
		if err != nil {
			return err
		}
		referenceAttempts = append(referenceAttempts, attempts...)
	}
	reference := references.String()
	var summaries []analysis.RunSummary
	var analyzed []analysis.RunAttempt
	for _, label := range conditions {
		attempts, err := load(label)
		if err != nil {
			return err
		}
		analyzed = append(analyzed, attempts...)
		if !*byModel {
			summaries = append(summaries, analysis.Summarize(label, attempts, referenceAttempts))
		}
	}
	var mean *analysis.ModelMean
	if *byModel {
		models, groups := analysis.ByModel(analyzed)
		_, referenceGroups := analysis.ByModel(referenceAttempts)
		for _, model := range models {
			summaries = append(summaries, analysis.Summarize(modelLabel(model), groups[model], referenceGroups[model]))
		}
		value := analysis.MeanOverModels(analyzed, referenceAttempts)
		mean = &value
	}
	for _, summary := range summaries {
		printRunSummary(output, summary, reference)
	}
	if mean != nil {
		printModelMean(output, summaries, *mean)
	}
	if *out == "" {
		return nil
	}
	if err := analysis.WriteRunAnalysis(*out, summaries, analyzed); err != nil {
		return err
	}
	if mean != nil {
		return analysis.WriteModelMean(*out, *mean)
	}
	return nil
}

func modelLabel(model string) string {
	if model == "" {
		return "unknown model"
	}
	return model
}

func printModelMean(output io.Writer, summaries []analysis.RunSummary, mean analysis.ModelMean) {
	table := tabwriter.NewWriter(output, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintf(table, "model\tattempts\tmacro\tfull\treference macro\tdifference\t95%% CI\t\n")
	for index, summary := range summaries {
		referenceMacro, difference, interval := "-", "-", "-"
		if value, ok := mean.ReferenceMacros[mean.Models[index]]; ok {
			referenceMacro = fmt.Sprintf("%.3f", value)
		}
		if d := summary.MacroDifference; d != nil {
			difference = fmt.Sprintf("%+.3f", d.Mean)
			interval = fmt.Sprintf("[%+.3f, %+.3f]", d.Low, d.High)
		}
		fmt.Fprintf(table, "%s\t%d\t%.3f\t%d/%d\t%s\t%s\t%s\t\n", summary.Label, summary.Attempts, summary.Macro, summary.FullSuccess, summary.Attempts, referenceMacro, difference, interval)
	}
	referenceMacro, difference, interval := "-", "-", "-"
	if mean.ReferenceMacro != nil {
		referenceMacro = fmt.Sprintf("%.3f", *mean.ReferenceMacro)
	}
	if d := mean.Difference; d != nil {
		difference = fmt.Sprintf("%+.3f", d.Mean)
		interval = fmt.Sprintf("[%+.3f, %+.3f]", d.Low, d.High)
	}
	fmt.Fprintf(table, "mean over %d models\t\t%.3f\t\t%s\t%s\t%s\t\n", len(mean.Models), mean.Macro, referenceMacro, difference, interval)
	_ = table.Flush()
	fmt.Fprintln(output)
}

func printRunSummary(output io.Writer, summary analysis.RunSummary, reference string) {
	fmt.Fprintf(output, "%s: %d attempts, macro %.3f, full success %d/%d, agent not run %d\n", summary.Label, summary.Attempts, summary.Macro, summary.FullSuccess, summary.Attempts, summary.AgentNotRun)
	if summary.Searches > 0 {
		fmt.Fprintf(output, "searches: %d in %d/%d attempts, %d before the first change, %d errors; source in top k: %d/%d searches, %d/%d attempts\n",
			summary.Searches, summary.SearchAttempts, summary.Attempts, summary.SearchesBeforeChange, summary.SearchErrors,
			summary.SourceHits, summary.Searches, summary.SourceAttempts, summary.Attempts)
	}
	if difference := summary.MacroDifference; difference != nil {
		fmt.Fprintf(output, "macro difference from %s on %d scenarios: %+.3f [95%% CI %+.3f, %+.3f]\n", reference, difference.Scenarios, difference.Mean, difference.Low, difference.High)
	}
	printUsage(output, summary)
	table := tabwriter.NewWriter(output, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintf(table, "group\tattempts\tscenarios\tmacro\tfull\treference macro\treference full\t\n")
	for _, group := range summary.Groups {
		referenceMacro, referenceFull := "-", "-"
		if group.ReferenceAvailable {
			referenceMacro = fmt.Sprintf("%.3f", group.ReferenceMacro)
			referenceFull = fmt.Sprintf("%d/%d", group.ReferenceFull, group.ReferenceAttempts)
		}
		fmt.Fprintf(table, "%s\t%d\t%d\t%.3f\t%d/%d\t%s\t%s\t\n", group.Name, group.Attempts, group.Scenarios, group.Macro, group.FullSuccess, group.Attempts, referenceMacro, referenceFull)
	}
	_ = table.Flush()
	fmt.Fprintln(output)
	table = tabwriter.NewWriter(output, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintf(table, "scenario\tscore\tfull\tsearched\tsearches\tsource found\t%s\t\n", referenceHeader(reference))
	for _, row := range summary.Scenarios {
		referenceScore := "-"
		if row.ReferenceScore != nil {
			referenceScore = fmt.Sprintf("%.2f", *row.ReferenceScore)
		}
		fmt.Fprintf(table, "%s\t%.2f\t%d/%d\t%d\t%d\t%d\t%s\t\n", row.Scenario, row.MeanScore, row.FullSuccess, row.Attempts, row.SearchAttempts, row.Searches, row.SourceRetrieved, referenceScore)
	}
	_ = table.Flush()
	fmt.Fprintln(output)
}

func printUsage(output io.Writer, summary analysis.RunSummary) {
	mean := func(label string, usage analysis.Usage) {
		fmt.Fprintf(output, "%s: %d agent attempts, mean %.1f turns, %.0f prompt tokens, %.0f completion tokens\n", label, usage.Attempts, usage.Turns, usage.Prompt, usage.Completion)
		fmt.Fprintf(output, "  peak context: mean %.0f, max %d tokens; %d attempts ran out of context\n", usage.PeakContext, usage.MaxPeakContext, usage.Overflows)
		fmt.Fprintf(output, "  %d kubectl changes, %d failed, %d kubectl edit; rollout status in %d attempts; %d completed without full success\n", usage.Changes, usage.FailedChanges, usage.Edits, usage.RolloutStatus, usage.Unconfirmed)
	}
	mean("usage", summary.Usage)
	if summary.ReferenceUsage != nil {
		mean("reference usage", *summary.ReferenceUsage)
	}
	fmt.Fprintln(output)
	table := tabwriter.NewWriter(output, 0, 0, 2, ' ', tabwriter.AlignRight)
	fmt.Fprintf(table, "scenario\tcriterion\tpassed\treference\t\n")
	for _, rate := range summary.Criteria {
		if rate.ReferenceAvailable && rate.Passed*rate.ReferenceTotal == rate.ReferencePassed*rate.Total {
			continue
		}
		reference := "-"
		if rate.ReferenceAvailable {
			reference = fmt.Sprintf("%d/%d", rate.ReferencePassed, rate.ReferenceTotal)
		}
		fmt.Fprintf(table, "%s\t%s\t%d/%d\t%s\t\n", rate.Scenario, rate.ID, rate.Passed, rate.Total, reference)
	}
	_ = table.Flush()
	fmt.Fprintln(output, "criteria listed where the pass rate differs from the reference (all in summary.json)")
	fmt.Fprintln(output)
}

func referenceHeader(reference string) string {
	if reference == "" {
		return "reference"
	}
	return reference
}
