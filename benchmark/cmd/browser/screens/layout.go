package screens

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
)

// At twoColumnWidth columns and wider, lists and the attempt details sit
// beside their preview or conversation pane.
const (
	twoColumnWidth = 112
	columnGap      = 2
)

func TwoColumn(width int) bool { return width >= twoColumnWidth }

func ColumnWidths(width int) (int, int) {
	left := (width - columnGap) * 58 / 100
	return left, width - columnGap - left
}

func PaneWidth(width int) int {
	if TwoColumn(width) {
		left, _ := ColumnWidths(width)
		return left
	}
	return width
}

func PreviewWidth(width int) int {
	if TwoColumn(width) {
		_, right := ColumnWidths(width)
		return right
	}
	return width
}

func ChatVisible(width, height int) int {
	if TwoColumn(width) {
		return max(1, height-1)
	}
	return max(1, height/2-1)
}

func DetailsHeight(width, height int) int {
	if TwoColumn(width) {
		return height
	}
	return max(1, height/2)
}

func Join(left, right string, width, height int) string {
	leftWidth, rightWidth := ColumnWidths(width)
	joined := lipgloss.JoinHorizontal(lipgloss.Top,
		padBlock(left, leftWidth), strings.Repeat(" ", columnGap), padBlock(right, rightWidth))
	return ui.FitHeight(joined, height)
}

func padBlock(content string, width int) string {
	lines := strings.Split(content, "\n")
	for index, line := range lines {
		lines[index] = ui.PadRight(line, width)
	}
	return strings.Join(lines, "\n")
}
