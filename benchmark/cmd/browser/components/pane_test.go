package components

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCardRowWrapsToWidth(t *testing.T) {
	cards := []string{"aaaa", "bbbb", "cccc"}
	assert.Equal(t, "aaaa  bbbb  cccc", CardRow(cards, 40))
	assert.Len(t, strings.Split(CardRow(cards, 10), "\n"), 2)
	assert.Len(t, strings.Split(CardRow(cards, 8), "\n"), 3)
	assert.Equal(t, "", CardRow(nil, 10))
}
