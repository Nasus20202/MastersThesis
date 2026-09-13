package logging

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"sync"
	"time"
)

const (
	colorReset         = "\x1b[0m"
	colorDim           = "\x1b[90m"
	colorCyan          = "\x1b[36m"
	colorGreen         = "\x1b[32m"
	colorYellow        = "\x1b[33m"
	colorRed           = "\x1b[31m"
	prettyMessageWidth = 32
	prettyTimeFormat   = "15:04:05"
	prettyAttrSpacing  = "  "
)

type prettyHandler struct {
	writer  io.Writer
	options slog.HandlerOptions
	color   bool
	mu      *sync.Mutex
	attrs   []prettyAttr
	groups  []string
}

type prettyAttr struct {
	groups []string
	attr   slog.Attr
}

type renderedAttr struct {
	key   string
	value slog.Value
}

func newPrettyHandler(writer io.Writer, options *slog.HandlerOptions, color bool) slog.Handler {
	var handlerOptions slog.HandlerOptions
	if options != nil {
		handlerOptions = *options
	}
	return &prettyHandler{
		writer:  writer,
		options: handlerOptions,
		color:   color,
		mu:      &sync.Mutex{},
	}
}

func (h *prettyHandler) Enabled(_ context.Context, level slog.Level) bool {
	return level >= h.options.Level.Level()
}

func (h *prettyHandler) Handle(_ context.Context, record slog.Record) error {
	attrs := make([]renderedAttr, 0, len(h.attrs)+record.NumAttrs())
	for _, attr := range h.attrs {
		appendRenderedAttrs(&attrs, attr.attr, attr.groups, h.options.ReplaceAttr)
	}
	record.Attrs(func(attr slog.Attr) bool {
		appendRenderedAttrs(&attrs, attr, h.groups, h.options.ReplaceAttr)
		return true
	})

	inline := make([]renderedAttr, 0, len(attrs))
	blocks := make([]renderedAttr, 0)
	for _, attr := range attrs {
		if isBlockAttribute(attr.key, attr.value) {
			blocks = append(blocks, attr)
			continue
		}
		if attr.value.Kind() == slog.KindString && attr.value.String() == "" {
			continue
		}
		inline = append(inline, attr)
	}

	var output strings.Builder
	if !record.Time.IsZero() {
		output.WriteString(colorize(record.Time.Format(prettyTimeFormat), colorDim, h.color))
		output.WriteByte(' ')
	}
	level := fmt.Sprintf("%-5s", formatLevel(record.Level))
	output.WriteString(colorize(level, levelColor(record.Level), h.color))
	output.WriteString(" | ")
	if len(inline) > 0 || len(blocks) > 0 {
		output.WriteString(fmt.Sprintf("%-*s", prettyMessageWidth, record.Message))
		output.WriteString(" |")
	} else {
		output.WriteString(record.Message)
	}
	for index, attr := range inline {
		if index == 0 {
			output.WriteByte(' ')
		} else {
			output.WriteString(prettyAttrSpacing)
		}
		output.WriteString(colorize(attr.key, colorDim, h.color))
		output.WriteString(colorize("=", colorDim, h.color))
		output.WriteString(formatInline(attr.value))
	}
	if h.options.AddSource && record.PC != 0 {
		if source := sourceLocation(record.PC); source != "" {
			if len(inline) == 0 {
				output.WriteByte(' ')
			} else {
				output.WriteString(prettyAttrSpacing)
			}
			output.WriteString(colorize("source", colorDim, h.color))
			output.WriteString(colorize("=", colorDim, h.color))
			output.WriteString(source)
		}
	}
	for _, attr := range blocks {
		output.WriteByte('\n')
		output.WriteString("  ")
		output.WriteString(colorize(attr.key, colorDim, h.color))
		output.WriteString(":\n")
		content := formatValue(attr.value)
		lines := strings.Split(content, "\n")
		for _, line := range lines {
			output.WriteString("    ")
			output.WriteString(line)
			output.WriteByte('\n')
		}
		if len(lines) == 0 {
			output.WriteString("    <empty>\n")
		}
	}
	if len(blocks) == 0 {
		output.WriteByte('\n')
	}

	h.mu.Lock()
	defer h.mu.Unlock()
	_, err := io.WriteString(h.writer, output.String())
	return err
}

func (h *prettyHandler) WithAttrs(attrs []slog.Attr) slog.Handler {
	if len(attrs) == 0 {
		return h
	}
	cloned := clonePrettyAttrs(h.attrs)
	for _, attr := range attrs {
		cloned = append(cloned, prettyAttr{groups: append([]string(nil), h.groups...), attr: attr})
	}
	return &prettyHandler{writer: h.writer, options: h.options, color: h.color, mu: h.mu, attrs: cloned, groups: append([]string(nil), h.groups...)}
}

func (h *prettyHandler) WithGroup(name string) slog.Handler {
	if name == "" {
		return h
	}
	return &prettyHandler{writer: h.writer, options: h.options, color: h.color, mu: h.mu, attrs: clonePrettyAttrs(h.attrs), groups: append(append([]string(nil), h.groups...), name)}
}

func appendRenderedAttrs(attrs *[]renderedAttr, attr slog.Attr, groups []string, replace func([]string, slog.Attr) slog.Attr) {
	attr.Value = attr.Value.Resolve()
	if replace != nil {
		attr = replace(groups, attr)
		attr.Value = attr.Value.Resolve()
	}
	if attr.Key == "" && attr.Value.Kind() == slog.KindGroup {
		for _, child := range attr.Value.Group() {
			appendRenderedAttrs(attrs, child, groups, replace)
		}
		return
	}
	if attr.Value.Kind() == slog.KindGroup {
		for _, child := range attr.Value.Group() {
			appendRenderedAttrs(attrs, child, append(append([]string(nil), groups...), attr.Key), replace)
		}
		return
	}
	key := attr.Key
	if len(groups) > 0 {
		key = strings.Join(append(append([]string(nil), groups...), key), ".")
	}
	if key != "" {
		*attrs = append(*attrs, renderedAttr{key: key, value: attr.Value})
	}
}

func clonePrettyAttrs(attrs []prettyAttr) []prettyAttr {
	cloned := make([]prettyAttr, len(attrs))
	copy(cloned, attrs)
	for index := range cloned {
		cloned[index].groups = append([]string(nil), cloned[index].groups...)
	}
	return cloned
}

func isBlockAttribute(_ string, value slog.Value) bool {
	return value.Kind() == slog.KindString && strings.Contains(value.String(), "\n")
}

func formatInline(value slog.Value) string {
	if value.Kind() == slog.KindString {
		return quoteIfNeeded(value.String())
	}
	return formatValue(value)
}

func formatValue(value slog.Value) string {
	switch value.Kind() {
	case slog.KindString:
		return value.String()
	case slog.KindTime:
		return value.Time().Format(time.RFC3339Nano)
	case slog.KindDuration:
		return value.Duration().String()
	case slog.KindAny:
		return fmt.Sprint(value.Any())
	default:
		return value.String()
	}
}

func quoteIfNeeded(value string) string {
	if value == "" || strings.ContainsAny(value, " \t=\"") {
		return strconv.Quote(value)
	}
	return value
}

func formatLevel(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return "ERROR"
	case level >= slog.LevelWarn:
		return "WARN"
	case level >= slog.LevelInfo:
		return "INFO"
	default:
		return "DEBUG"
	}
}

func levelColor(level slog.Level) string {
	switch {
	case level >= slog.LevelError:
		return colorRed
	case level >= slog.LevelWarn:
		return colorYellow
	case level >= slog.LevelInfo:
		return colorGreen
	default:
		return colorCyan
	}
}

func colorize(value, color string, enabled bool) string {
	if !enabled {
		return value
	}
	return color + value + colorReset
}

func colorEnabled(writer io.Writer) bool {
	file, ok := writer.(*os.File)
	if !ok {
		return false
	}
	info, err := file.Stat()
	return err == nil && info.Mode()&os.ModeCharDevice != 0
}

func sourceLocation(pc uintptr) string {
	frame, _ := runtime.CallersFrames([]uintptr{pc}).Next()
	if frame.File == "" {
		return ""
	}
	return filepath.Base(frame.File) + ":" + strconv.Itoa(frame.Line)
}
