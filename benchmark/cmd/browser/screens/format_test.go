package screens

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestShortModel(t *testing.T) {
	assert.Equal(t, "gemma-4-E4B", shortModel("gemma-4-E4B-it-qat-UD-Q4_K_XL"))
	assert.Equal(t, "Qwen3.5-9B", shortModel("Qwen3.5-9B-Q4_K_M"))
	assert.Equal(t, "plain", shortModel("plain"))
}
