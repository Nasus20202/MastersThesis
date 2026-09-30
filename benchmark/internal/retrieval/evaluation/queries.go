package evaluation

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

const QueriesPerProbe = 3

const queryInstruction = `Write %d short search queries for the Kubernetes documentation that would help troubleshoot the problem below. Use keywords or a short phrase, not full sentences. Reply with the %d queries only, one per line.

Problem:
%s`

// QuerySet is the frozen evaluation input with the settings that produced it.
type QuerySet struct {
	ProbesFile   string             `json:"probes_file"`
	ProbesSHA256 string             `json:"probes_sha256"`
	Instruction  string             `json:"instruction"`
	Model        inference.Metadata `json:"model"`
	Temperature  float64            `json:"temperature"`
	Seed         int                `json:"seed"`
	Probes       []ProbeQueries     `json:"probes"`
}

type ProbeQueries struct {
	Probe
	Queries   []string `json:"queries"`
	Response  string   `json:"response"`
	Reasoning string   `json:"reasoning,omitempty"`
}

func QueryInstruction() string {
	return fmt.Sprintf(queryInstruction, QueriesPerProbe, QueriesPerProbe, "{probe prompt}")
}

// GenerateQueries shows the model only the probe prompt, never its sources.
func GenerateQueries(ctx context.Context, client inference.Client, probes []Probe, temperature float64, seed int) ([]ProbeQueries, error) {
	generated := make([]ProbeQueries, 0, len(probes))
	for _, probe := range probes {
		prompt := fmt.Sprintf(queryInstruction, QueriesPerProbe, QueriesPerProbe, probe.Prompt)
		result, err := client.Chat(ctx, []inference.Message{{Role: "user", Content: prompt}}, nil,
			inference.Options{Temperature: &temperature, Seed: &seed})
		if err != nil {
			return nil, fmt.Errorf("generate queries for %s: %w", probe.ID, err)
		}
		queries, err := parseQueries(result.Message.Content)
		if err != nil {
			return nil, fmt.Errorf("generate queries for %s: %w", probe.ID, err)
		}
		slog.InfoContext(ctx, "generated queries", "probe", probe.ID, "queries", queries)
		generated = append(generated, ProbeQueries{
			Probe:     probe,
			Queries:   queries,
			Response:  result.Message.Content,
			Reasoning: result.Message.ReasoningContent,
		})
	}
	return generated, nil
}

var listMarker = regexp.MustCompile(`^(?:[-*•]|\d+[.)])\s*`)

// parseQueries keeps the first QueriesPerProbe lines without list markers or
// surrounding quotes.
func parseQueries(content string) ([]string, error) {
	var queries []string
	for _, line := range strings.Split(content, "\n") {
		query := listMarker.ReplaceAllString(strings.TrimSpace(line), "")
		query = strings.TrimSpace(strings.Trim(query, "\"'`*"))
		if query != "" {
			queries = append(queries, query)
		}
	}
	if len(queries) < QueriesPerProbe {
		return nil, fmt.Errorf("model returned %d queries, want %d: %q", len(queries), QueriesPerProbe, content)
	}
	return queries[:QueriesPerProbe], nil
}

// LoadQuerySet also returns the SHA-256 of the query set file.
func LoadQuerySet(path string) (QuerySet, string, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return QuerySet{}, "", err
	}
	var set QuerySet
	if err := json.Unmarshal(data, &set); err != nil {
		return QuerySet{}, "", fmt.Errorf("parse query set %s: %w", path, err)
	}
	return set, sha256Hex(data), nil
}

// WriteQuerySet refuses to replace an existing, frozen query set.
func WriteQuerySet(path string, set QuerySet) error {
	if _, err := os.Stat(path); err == nil {
		return fmt.Errorf("%s already exists; the query set is frozen once written", path)
	}
	if err := os.MkdirAll(filepath.Dir(path), 0o750); err != nil {
		return err
	}
	return writeJSON(path, set)
}
