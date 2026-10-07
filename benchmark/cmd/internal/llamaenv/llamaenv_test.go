package llamaenv

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParallelism(t *testing.T) {
	t.Run("configured", func(t *testing.T) {
		t.Setenv(envLlamaParallel, "2")
		parallelism, err := Parallelism()
		require.NoError(t, err)
		assert.Equal(t, 2, parallelism)
	})

	t.Run("compose default", func(t *testing.T) {
		t.Setenv(envLlamaParallel, "")
		parallelism, err := Parallelism()
		require.NoError(t, err)
		assert.Equal(t, 1, parallelism)
	})

	for _, value := range []string{"0", "-1", "many"} {
		t.Run("invalid "+value, func(t *testing.T) {
			t.Setenv(envLlamaParallel, value)
			_, err := Parallelism()
			assert.ErrorContains(t, err, envLlamaParallel+" must be a positive integer")
		})
	}
}

func TestModelArtifactAndRuntimeSettings(t *testing.T) {
	t.Setenv(envLlamaModelRepository, "google/gemma")
	t.Setenv(envLlamaModelRevision, "revision")
	t.Setenv(envLlamaModelFile, "gemma.gguf")
	t.Setenv(envLlamaModelQuant, "Q4_0")
	t.Setenv(envLlamaModelSHA256, "hash")
	t.Setenv(envLlamaKVUnified, "32768")
	t.Setenv(envLlamaFlashAttention, "auto")
	t.Setenv(envLlamaReasoning, "on")
	t.Setenv(envLlamaSpecType, "draft-mtp")
	t.Setenv(envLlamaSpecDraftNMax, "3")
	t.Setenv(envLlamaDraftRepository, "unsloth/gemma")
	t.Setenv(envLlamaDraftRevision, "draft-revision")
	t.Setenv(envLlamaDraftFile, "MTP/drafter.gguf")
	t.Setenv(envLlamaDraftSHA256, "draft-hash")

	assert.Equal(t, "google/gemma@revision/gemma.gguf", artifact(envLlamaModelRepository, envLlamaModelRevision, envLlamaModelFile))
	settings := runtimeSettings()
	assert.Equal(t, "32768", settings[envLlamaKVUnified])
	assert.Equal(t, "on", settings[envLlamaReasoning])
	assert.Equal(t, "auto", settings[envLlamaFlashAttention])
	assert.Equal(t, "draft-mtp", settings[envLlamaSpecType])
	assert.Equal(t, "3", settings[envLlamaSpecDraftNMax])
	assert.Equal(t, "unsloth/gemma", settings[envLlamaDraftRepository])
	assert.Equal(t, "draft-revision", settings[envLlamaDraftRevision])
	assert.Equal(t, "MTP/drafter.gguf", settings[envLlamaDraftFile])
	assert.Equal(t, "draft-hash", settings[envLlamaDraftSHA256])
}

func TestNewEmbeddingClientRecordsArtifact(t *testing.T) {
	t.Setenv(envEmbeddingModelName, "")
	_, err := NewEmbeddingClient()
	assert.ErrorContains(t, err, envEmbeddingModelName+" and "+envEmbeddingPort+" or "+envEmbeddingBaseURL+" are required")

	t.Setenv(envEmbeddingModelName, "embeddinggemma")
	t.Setenv(envEmbeddingPort, "")
	t.Setenv(envEmbeddingBaseURL, "http://127.0.0.1:8090")
	_, err = NewEmbeddingClient()
	require.NoError(t, err)

	t.Setenv(envEmbeddingModelName, "embeddinggemma")
	t.Setenv(envEmbeddingPort, "8081")
	t.Setenv(envEmbeddingRepository, "ggml-org/embeddinggemma")
	t.Setenv(envEmbeddingRevision, "revision")
	t.Setenv(envEmbeddingFile, "model.gguf")
	client, err := NewEmbeddingClient()
	require.NoError(t, err)
	assert.Equal(t, "ggml-org/embeddinggemma@revision/model.gguf", client.Metadata().Artifact)
}

func TestNewChatClientRecordsRuntimeSettingsOnlyForLocalServer(t *testing.T) {
	t.Setenv(envLlamaModelName, "gemma")
	t.Setenv(envLlamaKVUnified, "32768")

	client, err := NewChatClient()
	require.NoError(t, err)
	assert.Equal(t, "32768", client.Metadata().RuntimeSettings[envLlamaKVUnified])

	t.Setenv(envLlamaBaseURL, "http://127.0.0.1:8090")
	client, err = NewChatClient()
	require.NoError(t, err)
	assert.Nil(t, client.Metadata().RuntimeSettings)
}
