// Table renders the browser's responsive, selectable data tables. It is
// presentational: the owning screen keeps the cursor and scroll offset so click
// handling and keyboard navigation stay in one place.
package components

import (
	"strings"

	"charm.land/lipgloss/v2"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
)

// Column is one responsive column: it shrinks to Min and grows to Max.
type Column struct {
	Title string
	Min   int
	Max   int
}

// Table renders rows against a fixed set of columns.
type Table struct {
	Columns []Column
}

// New returns a table with the given columns.
func NewTable(columns ...Column) Table {
	return Table{Columns: columns}
}

// Render draws the header and the visible window of rows. Rows at Cursor get
// the selected style; other rows are colored by styleFor when provided.
func (t Table) Render(width, height int, rows [][]string, cursor, offset int, styleFor func(int) lipgloss.Style) string {
	widths := t.columnWidths(rows, max(1, width-2))
	headers := make([]string, len(t.Columns))
	for index, column := range t.Columns {
		headers[index] = column.Title
	}
	var builder strings.Builder
	builder.WriteString(ui.TableHeader.Width(max(1, width)).Render("  " + formatRow(headers, widths)))
	builder.WriteByte('\n')
	if len(rows) == 0 {
		builder.WriteString(ui.MutedStyle.Render("  no entries"))
		builder.WriteByte('\n')
		return builder.String()
	}
	visible := max(1, height-1)
	offset = ui.ClampOffset(offset, len(rows), visible)
	end := min(len(rows), offset+visible)
	for index := offset; index < end; index++ {
		gutter := "  "
		if index == cursor {
			gutter = ui.Gutter.Render("▸ ")
		}
		line := gutter + formatRow(rows[index], widths)
		switch {
		case index == cursor:
			line = ui.SelectedRow.Width(max(1, width)).Render(line)
		case styleFor != nil:
			line = styleFor(index).Render(line)
		}
		builder.WriteString(line)
		builder.WriteByte('\n')
	}
	return builder.String()
}

// VisibleRows reports how many data rows fit in a table of the given height.
func VisibleRows(height int) int {
	return max(1, height-1)
}

func formatRow(cells []string, widths []int) string {
	parts := make([]string, len(widths))
	for index := range widths {
		cell := ""
		if index < len(cells) {
			cell = cells[index]
		}
		parts[index] = ui.PadRight(ui.Truncate(cell, widths[index]), widths[index])
	}
	return strings.Join(parts, " ")
}

func (t Table) columnWidths(rows [][]string, width int) []int {
	widths := make([]int, len(t.Columns))
	for index, column := range t.Columns {
		widths[index] = lipgloss.Width(column.Title)
	}
	for _, row := range rows {
		for index := range t.Columns {
			if index >= len(row) {
				continue
			}
			if cell := lipgloss.Width(row[index]); cell > widths[index] {
				widths[index] = cell
			}
		}
	}
	for index, column := range t.Columns {
		widths[index] = min(widths[index], column.Max)
		widths[index] = max(widths[index], column.Min)
	}
	total := len(t.Columns) - 1
	for _, column := range widths {
		total += column
	}
	if total < width {
		widths[len(widths)-1] += width - total
		return widths
	}
	for total > width {
		widest, slack := -1, 0
		for index := range widths {
			if candidate := widths[index] - t.Columns[index].Min; candidate > slack {
				slack, widest = candidate, index
			}
		}
		if widest < 0 {
			break
		}
		widths[widest]--
		total--
	}
	return widths
}
