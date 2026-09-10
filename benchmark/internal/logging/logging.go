package logging

import (
	"errors"
	"fmt"
	"io"
	"log/slog"
	"strings"
)

type Format string

const (
	FormatText Format = "text"
	FormatJSON Format = "json"
)

func New(w io.Writer, format Format, level slog.Level) (*slog.Logger, error) {
	if w == nil {
		return nil, errors.New("log writer is required")
	}

	options := &slog.HandlerOptions{
		AddSource: level <= slog.LevelDebug,
		Level:     level,
	}

	switch strings.ToLower(strings.TrimSpace(string(format))) {
	case "", string(FormatText):
		return slog.New(slog.NewTextHandler(w, options)), nil
	case string(FormatJSON):
		return slog.New(slog.NewJSONHandler(w, options)), nil
	default:
		return nil, fmt.Errorf("unsupported log format %q", format)
	}
}

func ParseLevel(value string) (slog.Level, error) {
	value = strings.TrimSpace(value)
	if value == "" {
		return slog.LevelInfo, nil
	}

	var level slog.Level
	if err := level.UnmarshalText([]byte(value)); err != nil {
		return 0, fmt.Errorf("parse log level %q: %w", value, err)
	}
	return level, nil
}
