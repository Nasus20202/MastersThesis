package main

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func writeRecord(t *testing.T, body string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "001.json")
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	return path
}

func TestLoadPrefixStopsBeforeTheFinalAssistantMessage(t *testing.T) {
	path := writeRecord(t, `{"scenario_id":"pvc-pending","agent":{"messages":[
		{"role":"system","content":"s"},{"role":"user","content":"u"},
		{"role":"assistant","content":"","tool_calls":[{"id":"c","type":"function","name":"bash","arguments":"{}"}]},
		{"role":"tool","tool_call_id":"c","content":"out"},
		{"role":"assistant","content":"done"}]}}`)

	prefix, scenario, err := loadPrefix("/repo", path)

	require.NoError(t, err)
	assert.Equal(t, "pvc-pending", scenario)
	assert.Equal(t, path, prefix.Source)
	require.Len(t, prefix.Messages, 4)
	assert.Equal(t, "tool", prefix.Messages[3].Role)
}

func TestLoadPrefixRejectsARecordThatDidNotEndOnAFinalMessage(t *testing.T) {
	path := writeRecord(t, `{"scenario_id":"s","agent":{"messages":[
		{"role":"system","content":"s"},{"role":"user","content":"u"},
		{"role":"assistant","content":"","tool_calls":[{"id":"c","type":"function","name":"bash","arguments":"{}"}]}]}}`)

	_, _, err := loadPrefix("", path)

	assert.ErrorContains(t, err, "final assistant message")
}

func TestRecordedSourceIsRelativeToTheRepository(t *testing.T) {
	root := t.TempDir()
	inside := filepath.Join(root, "docs", "run", "001.json")
	require.NoError(t, os.MkdirAll(filepath.Dir(inside), 0o700))
	require.NoError(t, os.WriteFile(inside, []byte("{}"), 0o600))
	elsewhere := writeRecord(t, "{}")

	assert.Equal(t, "docs/run/001.json", recordedSource(root, inside))
	assert.Equal(t, elsewhere, recordedSource(root, elsewhere))
	assert.Equal(t, inside, recordedSource("", inside))
}

func writePrefix(t *testing.T, name, scenario string) string {
	t.Helper()
	body := fmt.Sprintf(`{"scenario_id":%q,"agent":{"messages":[
		{"role":"system","content":"s"},{"role":"user","content":"u"},
		{"role":"assistant","content":"done"}]}}`, scenario)
	path := filepath.Join(t.TempDir(), name)
	require.NoError(t, os.WriteFile(path, []byte(body), 0o600))
	return path
}

func TestLoadPrefixesKeepsOnePrefixPerScenario(t *testing.T) {
	pvc := writePrefix(t, "a.json", "pvc-pending")
	oom := writePrefix(t, "b.json", "oom-memory-limit")

	prefixes, err := loadPrefixes([]string{pvc, oom})

	require.NoError(t, err)
	require.Len(t, prefixes, 2)
	assert.Equal(t, pvc, prefixes["pvc-pending"].Source)
	assert.Equal(t, oom, prefixes["oom-memory-limit"].Source)
}

func TestLoadPrefixesRejectsTwoRecordsForOneScenario(t *testing.T) {
	first := writePrefix(t, "a.json", "pvc-pending")
	second := writePrefix(t, "b.json", "pvc-pending")

	_, err := loadPrefixes([]string{first, second})

	assert.ErrorContains(t, err, `two recorded attempts for scenario "pvc-pending"`)
}
