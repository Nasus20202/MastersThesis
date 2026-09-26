package main

import (
	"context"
	"os/exec"
	"strings"
)

func repositoryProvenance(ctx context.Context) (string, bool) {
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
