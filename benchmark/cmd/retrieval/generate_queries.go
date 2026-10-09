package main

import (
	"context"
	"errors"
	"flag"
	"io"
	"path/filepath"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/internal/inferenceenv"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/retrieval/evaluation"
)

// Query generation settings fixed by the study design.
const (
	queryTemperature = 0.0
	querySeed        = 42
)

func runGenerateQueries(ctx context.Context, flags *flag.FlagSet, args []string, _ io.Writer) error {
	probes := flags.String("probes", "", "knowledge-check task file with the probes")
	out := flags.String("out", "", "query set file to write")
	if err := parse(flags, args); err != nil {
		return err
	}
	return generateQueries(ctx, *probes, *out)
}

func generateQueries(ctx context.Context, probesPath, outPath string) error {
	if probesPath == "" || outPath == "" {
		return errors.New("--probes and --out are required")
	}
	probes, probesSHA256, err := evaluation.LoadProbes(probesPath)
	if err != nil {
		return err
	}
	client, err := inferenceenv.NewChatClient()
	if err != nil {
		return err
	}
	generated, err := evaluation.GenerateQueries(ctx, client, probes, queryTemperature, querySeed)
	if err != nil {
		return err
	}
	set := evaluation.QuerySet{
		ProbesFile:   filepath.ToSlash(probesPath),
		ProbesSHA256: probesSHA256,
		Instruction:  evaluation.QueryInstruction(),
		Model:        client.Metadata(),
		Temperature:  queryTemperature,
		Seed:         querySeed,
		Probes:       generated,
	}
	return evaluation.WriteQuerySet(outPath, set)
}
