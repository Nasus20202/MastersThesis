package screens

import (
	"strings"

	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/components"
	"github.com/Nasus20202/MastersThesis/benchmark/cmd/browser/ui"
)

// List is the primary pane for list routes.
func (v *View) List(route *Route, width, height int) string {
	contentWidth := max(1, PaneWidth(width)-1)
	head := v.screenHead(route, contentWidth)
	tableHeight := max(1, height-ui.LineCount(head))
	visible := max(1, tableHeight-1)
	body := head + "\n" + v.screenTable(route, contentWidth, tableHeight)
	return components.Frame(body, contentWidth, height, v.ListCount(route), route.Offset, visible)
}

// Right is the secondary pane: a scrollable dashboard or preview.
func (v *View) Right(route *Route, width, height int) string {
	contentWidth := max(1, PreviewWidth(width)-1)
	key := previewKey(route, v.mode, contentWidth)
	lines, ok := v.previews[key]
	if !ok {
		lines = strings.Split(v.previewContent(route, contentWidth), "\n")
		v.previews[key] = lines
	}
	view := components.NewViewport()
	view.SetSize(contentWidth, height)
	view.SetLines(lines)
	view.SetOffset(route.PreviewOffset)
	return view.View()
}

// PreviewLines returns the un-windowed preview content for scrolling.
func (v *View) PreviewLines(route *Route, width int) []string {
	return v.previewLines(route, max(1, PreviewWidth(width)-1))
}
