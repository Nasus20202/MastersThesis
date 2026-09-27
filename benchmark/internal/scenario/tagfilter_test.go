package scenario

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseTagFilter(t *testing.T) {
	filter, err := ParseTagFilter([]string{"difficulty=hard", "area=networking", "difficulty=easy"})
	require.NoError(t, err)
	assert.Equal(t, TagFilter{"difficulty": {"hard", "easy"}, "area": {"networking"}}, filter)
	assert.Equal(t, "area=networking,difficulty=easy|hard", filter.String())

	_, err = ParseTagFilter([]string{"difficulty"})
	assert.Error(t, err)

	empty, err := ParseTagFilter([]string{"", "  "})
	require.NoError(t, err)
	assert.True(t, empty.Empty())
	assert.True(t, empty.Matches(map[string]string{}))
}

func TestTagFilterMatches(t *testing.T) {
	tags := map[string]string{"difficulty": "hard", "area": "networking"}
	assert.True(t, TagFilter{"difficulty": {"hard"}}.Matches(tags))
	assert.True(t, TagFilter{"difficulty": {"easy", "hard"}}.Matches(tags))
	assert.True(t, TagFilter{"difficulty": {"hard"}, "area": {"networking"}}.Matches(tags))
	assert.False(t, TagFilter{"difficulty": {"easy"}}.Matches(tags))
	assert.False(t, TagFilter{"difficulty": {"hard"}, "area": {"pods"}}.Matches(tags))
	assert.False(t, TagFilter{"knowledge": {"general"}}.Matches(tags))
}

func TestFilterByTags(t *testing.T) {
	definitions := []Definition{
		{ID: "easy-task", Tags: map[string]string{"difficulty": "easy", "area": "pods"}},
		{ID: "hard-task", Tags: map[string]string{"difficulty": "hard", "area": "networking"}},
		{ID: "untagged"},
	}

	filtered := FilterByTags(definitions, TagFilter{"difficulty": {"hard"}})
	require.Len(t, filtered, 1)
	assert.Equal(t, "hard-task", filtered[0].ID)

	assert.Len(t, FilterByTags(definitions, TagFilter{"difficulty": {"easy", "hard"}}), 2)
	assert.Empty(t, FilterByTags(definitions, TagFilter{"difficulty": {"easy"}, "area": {"networking"}}))
	assert.Len(t, FilterByTags(definitions, TagFilter{}), 3)
}
