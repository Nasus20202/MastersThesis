package ui

import (
	"sort"
	"strings"
)

// Rank is one row in a ranked bar list.
type Rank struct {
	Label  string
	Value  float64 // 0..1
	Detail string
}

// Ranked renders labels with horizontal meters, best first. It is used where
// vertical bar charts would render long names as unreadable stacked letters.
func Ranked(ranks []Rank, width, limit int) string {
	if len(ranks) == 0 {
		return MutedStyle.Render("no data")
	}
	sort.SliceStable(ranks, func(i, j int) bool { return ranks[i].Value > ranks[j].Value })
	if limit > 0 && len(ranks) > limit {
		ranks = ranks[:limit]
	}
	labelWidth := min(30, max(12, width/3))
	barWidth := max(8, width-labelWidth-8)
	lines := make([]string, 0, len(ranks))
	for _, rank := range ranks {
		lines = append(lines, PadRight(Truncate(rank.Label, labelWidth), labelWidth)+
			Meter(rank.Value, barWidth, Outcome(rank.Value >= 1, rank.Value))+
			MutedStyle.Render(" "+Rate(rank.Value)+rank.Detail))
	}
	return strings.Join(lines, "\n")
}
