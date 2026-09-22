// Package ui holds the browser's rendering primitives: charts, markdown,
// text/layout helpers, the design system and key bindings.
package ui

import (
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/NimbleMarkets/ntcharts/v2/barchart"
	"github.com/NimbleMarkets/ntcharts/v2/sparkline"
)

// Item is one bar in a chart. Values are in the range 0..1.
type Item struct {
	Label string
	Value float64
	Style lipgloss.Style
}

// Bars renders a vertical bar chart of values in the range 0..1.
func Bars(width, height int, items []Item) string {
	if len(items) == 0 {
		return MutedStyle.Render("no data")
	}
	data := make([]barchart.BarData, 0, len(items))
	for _, item := range items {
		data = append(data, barchart.BarData{
			Label:  item.Label,
			Values: []barchart.BarValue{{Value: item.Value, Style: item.Style}},
		})
	}
	chart := barchart.New(max(10, width), max(3, height), barchart.WithDataSet(data), barchart.WithMaxValue(1))
	chart.Draw()
	return chart.View()
}

// Histogram renders count buckets scaled to the largest bucket.
func Histogram(width, height int, labels []string, counts []int) string {
	maximum := 0
	for _, count := range counts {
		maximum = max(maximum, count)
	}
	if maximum == 0 {
		return MutedStyle.Render("no data")
	}
	items := make([]Item, 0, len(counts))
	for index, count := range counts {
		label := ""
		if index < len(labels) {
			label = labels[index]
		}
		items = append(items, Item{Label: label, Value: float64(count) / float64(maximum), Style: AccentStyle})
	}
	return Bars(width, height, items)
}

// Sparkline renders a 0..1 time series.
func Sparkline(width, height int, values []float64, style lipgloss.Style) string {
	if len(values) == 0 {
		return MutedStyle.Render("no attempts")
	}
	chart := sparkline.New(max(10, width), max(2, height), sparkline.WithData(values), sparkline.WithMaxValue(1), sparkline.WithStyle(style))
	chart.Draw()
	return chart.View()
}

// Meter renders a single horizontal 0..1 bar of the given width.
func Meter(value float64, width int, style lipgloss.Style) string {
	width = max(1, width)
	value = max(0, min(1, value))
	filled := int(value*float64(width) + 0.5)
	return style.Render(strings.Repeat("█", filled)) + MutedStyle.Render(strings.Repeat("░", width-filled))
}

// Segment is one slice of a donut chart.
type Segment struct {
	Label string
	Value float64
	Style lipgloss.Style
}

// Donut renders a ring chart with the segments in order, plus a legend column.
func Donut(width, height int, segments []Segment) string {
	total := 0.0
	for _, segment := range segments {
		total += segment.Value
	}
	if total <= 0 {
		return MutedStyle.Render("no attempts")
	}
	ringWidth := max(6, width/2)
	height = max(3, height)
	cx := float64(ringWidth-1) / 2
	cy := float64(height-1) / 2
	rx := max(1.0, float64(ringWidth)/2)
	ry := max(1.0, float64(height)/2)

	var ring strings.Builder
	for y := 0; y < height; y++ {
		for x := 0; x < ringWidth; x++ {
			nx := (float64(x) - cx) / rx
			ny := (float64(y) - cy) / ry
			radius := math.Hypot(nx, ny)
			if radius > 1 || radius < 0.45 {
				ring.WriteString(" ")
				continue
			}
			angle := math.Atan2(ny, nx)
			fraction := (angle + math.Pi/2) / (2 * math.Pi)
			if fraction < 0 {
				fraction++
			}
			ring.WriteString(segmentAt(segments, total, fraction).Style.Render("█"))
		}
		ring.WriteString("\n")
	}

	legend := make([]string, 0, len(segments))
	for _, segment := range segments {
		legend = append(legend, segment.Style.Render("█")+" "+
			PadRight(segment.Label, 8)+
			MutedStyle.Render(Rate(segment.Value/total)+" ("+Float(segment.Value)+")"))
	}
	return lipgloss.JoinHorizontal(lipgloss.Top,
		strings.TrimRight(ring.String(), "\n"),
		"  ",
		strings.Join(legend, "\n"))
}

func segmentAt(segments []Segment, total, fraction float64) Segment {
	accumulated := 0.0
	for _, segment := range segments {
		accumulated += segment.Value / total
		if fraction <= accumulated {
			return segment
		}
	}
	return segments[len(segments)-1]
}
