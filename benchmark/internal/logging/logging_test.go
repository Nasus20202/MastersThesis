package logging

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

func TestNewCreatesConfiguredHandler(t *testing.T) {
	var output bytes.Buffer
	logger, err := New(&output, FormatJSON, slog.LevelInfo)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logger.Debug("hidden")
	logger.Info("visible", "component", "test")

	if strings.Contains(output.String(), "hidden") {
		t.Fatal("debug record was not filtered")
	}
	if !strings.Contains(output.String(), `"msg":"visible"`) {
		t.Fatalf("output = %q, want visible JSON record", output.String())
	}
	if strings.Contains(output.String(), `"source"`) {
		t.Fatalf("output = %q, did not expect source at info level", output.String())
	}
}

func TestNewAddsSourceAtDebugLevel(t *testing.T) {
	var output bytes.Buffer
	logger, err := New(&output, FormatJSON, slog.LevelDebug)
	if err != nil {
		t.Fatalf("New() error = %v", err)
	}

	logger.Debug("visible")

	if !strings.Contains(output.String(), `"source"`) {
		t.Fatalf("output = %q, want source at debug level", output.String())
	}
}

func TestNewRejectsUnknownFormat(t *testing.T) {
	if _, err := New(&bytes.Buffer{}, Format("xml"), slog.LevelInfo); err == nil {
		t.Fatal("expected unsupported format error")
	}
}

func TestParseLevel(t *testing.T) {
	level, err := ParseLevel(" WARN ")
	if err != nil {
		t.Fatalf("ParseLevel() error = %v", err)
	}
	if level != slog.LevelWarn {
		t.Fatalf("ParseLevel() = %v, want %v", level, slog.LevelWarn)
	}

	if _, err := ParseLevel("trace"); err == nil {
		t.Fatal("expected unsupported level error")
	}
	level, err = ParseLevel("")
	if err != nil {
		t.Fatalf("ParseLevel() empty error = %v", err)
	}
	if level != slog.LevelInfo {
		t.Fatalf("ParseLevel() empty = %v, want %v", level, slog.LevelInfo)
	}
}
