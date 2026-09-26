package ui

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCountAddsThousandsSeparators(t *testing.T) {
	assert.Equal(t, "0", Count(0))
	assert.Equal(t, "999", Count(999))
	assert.Equal(t, "1,000", Count(1000))
	assert.Equal(t, "1,234,567", Count(1234567))
	assert.Equal(t, "-1,234", Count(-1234))
}
