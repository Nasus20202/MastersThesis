// Small string and layout helpers shared by the browser
// components: padding, truncation, wrapping and line windowing.
package ui

import (
	"fmt"
	"strings"
	"time"

	"charm.land/lipgloss/v2"
)

// PadRight pads value with spaces to width columns. ANSI sequences are
// measured, not counted.
func PadRight(value string, width int) string {
	if padding := width - lipgloss.Width(value); padding > 0 {
		return value + strings.Repeat(" ", padding)
	}
	return value
}

// Truncate shortens value to width columns with an ellipsis, measuring ANSI.
func Truncate(value string, width int) string {
	if width <= 0 {
		return ""
	}
	if lipgloss.Width(value) <= width {
		return value
	}
	var builder strings.Builder
	used := 0
	for _, character := range value {
		characterWidth := lipgloss.Width(string(character))
		if used+characterWidth > width-1 {
			break
		}
		builder.WriteRune(character)
		used += characterWidth
	}
	return builder.String() + "…"
}

// TruncatePlain is like Truncate but for plain (unstyled) runes.
func TruncatePlain(value string, width int) string {
	runes := []rune(value)
	if len(runes) <= width {
		return value
	}
	if width <= 1 {
		return string(runes[:max(0, width)])
	}
	return string(runes[:width-1]) + "…"
}

// LineCount returns the number of lines in value.
func LineCount(value string) int {
	if value == "" {
		return 0
	}
	return strings.Count(value, "\n") + 1
}

// FitHeight truncates or pads content to exactly height lines.
func FitHeight(content string, height int) string {
	lines := strings.Split(content, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for len(lines) < height {
		lines = append(lines, "")
	}
	return strings.Join(lines, "\n")
}

// Window returns lines[offset:offset+height], clamping the offset.
func Window(lines []string, offset, height int) string {
	if len(lines) == 0 {
		return ""
	}
	height = max(1, height)
	offset = max(0, min(offset, max(0, len(lines)-height)))
	end := min(len(lines), offset+height)
	return strings.Join(lines[offset:end], "\n")
}

// ClampOffset clamps a scroll offset to the valid range for count lines.
func ClampOffset(offset, count, visible int) int {
	maxOffset := max(0, count-visible)
	return max(0, min(offset, maxOffset))
}

// JoinSides places right at the end of a width-wide line.
func JoinSides(left, right string, width int) string {
	gap := width - lipgloss.Width(left) - lipgloss.Width(right)
	if gap < 1 {
		return left + " " + right
	}
	return left + strings.Repeat(" ", gap) + right
}

// WrapPreview wraps text to width and clamps it to maxLines with a remainder
// marker.
func WrapPreview(text string, width, maxLines int) []string {
	wrapped := strings.Split(lipgloss.Wrap(strings.TrimSpace(text), max(10, width), ""), "\n")
	if len(wrapped) > maxLines {
		remaining := len(wrapped) - maxLines
		wrapped = append(wrapped[:maxLines], fmt.Sprintf("… %d more lines", remaining))
	}
	return wrapped
}

// Rate formats a 0..1 value as a whole percentage.
func Rate(value float64) string {
	return fmt.Sprintf("%.0f%%", value*100)
}

// Seconds formats a duration in seconds, or a dash when unset.
func Seconds(seconds float64) string {
	if seconds <= 0 {
		return "–"
	}
	return time.Duration(seconds * float64(time.Second)).Round(time.Second).String()
}

// Time formats a run timestamp, or a dash when zero.
func Time(value time.Time) string {
	if value.IsZero() {
		return "–"
	}
	return value.Local().Format("2006-01-02 15:04")
}

// Float trims trailing zeros from a float.
func Float(value float64) string {
	return strings.TrimRight(strings.TrimRight(fmt.Sprintf("%.2f", value), "0"), ".")
}

// FirstLine returns the first line of value.
func FirstLine(value string) string {
	if index := strings.IndexByte(value, '\n'); index >= 0 {
		return value[:index]
	}
	return value
}

// ShortRevision abbreviates a git revision.
func ShortRevision(revision string) string {
	if len(revision) > 8 {
		return revision[:8]
	}
	if revision == "" {
		return "–"
	}
	return revision
}

// YesNo renders a boolean as yes/no.
func YesNo(value bool) string {
	if value {
		return "yes"
	}
	return "no"
}
