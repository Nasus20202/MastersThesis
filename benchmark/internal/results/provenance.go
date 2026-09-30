package results

import (
	"context"
	"os/exec"
	"strings"
	"time"
)

// NewRunID names a run by its UTC start time with millisecond precision.
func NewRunID(startedAt time.Time) string {
	timestamp := startedAt.UTC().Format("2006-01-02-15-04-05.000Z")
	return "run-" + strings.Replace(timestamp, ".", "-", 1)
}

// RepositoryProvenance returns the current Git revision and whether the
// working tree has uncommitted changes.
func RepositoryProvenance(ctx context.Context) (string, bool) {
	rootOutput, err := exec.CommandContext(ctx, "git", "rev-parse", "--show-toplevel").Output()
	if err != nil {
		return "", false
	}
	root := strings.TrimSpace(string(rootOutput))
	if root == "" {
		return "", false
	}
	revisionOutput, err := exec.CommandContext(ctx, "git", "-C", root, "rev-parse", "HEAD").Output()
	if err != nil {
		return "", false
	}
	statusOutput, err := exec.CommandContext(ctx, "git", "-C", root, "status", "--porcelain").Output()
	if err != nil {
		return strings.TrimSpace(string(revisionOutput)), false
	}
	return strings.TrimSpace(string(revisionOutput)), strings.TrimSpace(string(statusOutput)) != ""
}
