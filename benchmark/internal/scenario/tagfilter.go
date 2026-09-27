package scenario

import (
	"fmt"
	"slices"
	"sort"
	"strings"
)

// TagFilter selects scenarios by tag value: AND across keys, OR within a key.
type TagFilter map[string][]string

func ParseTagFilter(values []string) (TagFilter, error) {
	filter := TagFilter{}
	for _, value := range values {
		for _, requirement := range strings.Split(value, ",") {
			requirement = strings.TrimSpace(requirement)
			if requirement == "" {
				continue
			}
			key, tagValue, ok := strings.Cut(requirement, "=")
			key = strings.TrimSpace(key)
			tagValue = strings.TrimSpace(tagValue)
			if !ok || key == "" || tagValue == "" {
				return nil, fmt.Errorf("invalid tag %q, want key=value", requirement)
			}
			if !slices.Contains(filter[key], tagValue) {
				filter[key] = append(filter[key], tagValue)
			}
		}
	}
	return filter, nil
}

func (f TagFilter) Empty() bool { return len(f) == 0 }

func (f TagFilter) Matches(tags map[string]string) bool {
	for key, values := range f {
		if !slices.Contains(values, tags[key]) {
			return false
		}
	}
	return true
}

func (f TagFilter) String() string {
	if len(f) == 0 {
		return ""
	}
	keys := make([]string, 0, len(f))
	for key := range f {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	parts := make([]string, 0, len(keys))
	for _, key := range keys {
		values := slices.Clone(f[key])
		sort.Strings(values)
		parts = append(parts, key+"="+strings.Join(values, "|"))
	}
	return strings.Join(parts, ",")
}

func FilterByTags(definitions []Definition, filter TagFilter) []Definition {
	if filter.Empty() {
		return definitions
	}
	filtered := make([]Definition, 0, len(definitions))
	for _, definition := range definitions {
		if filter.Matches(definition.Tags) {
			filtered = append(filtered, definition)
		}
	}
	return filtered
}
