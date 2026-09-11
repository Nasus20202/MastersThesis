package main

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestRunIDUsesReadableUTCNanosecondTimestamp(t *testing.T) {
	startedAt := time.Date(2026, time.September, 11, 11, 44, 22, 123456789, time.FixedZone("CEST", 2*60*60))

	assert.Equal(t, "run-2026-09-11-09-44-22-123Z", runID(startedAt))
}
