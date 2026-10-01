package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	commandagent "github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

func TestValidateResumeMetadataComparesRetrieval(t *testing.T) {
	recorded := results.RetrievalProvenance{Mode: "hybrid", Chunking: "windows", TopK: 5, MaxBytes: 8192, IndexSHA256: "a"}
	metadata := results.RunMetadata{RunType: results.RunTypeBenchmark, Agents: []string{"rag"}, RepeatCount: 1, Scenarios: []string{"s"}, Retrieval: &recorded}
	agents := []commandagent.Name{commandagent.RAG}

	same := recorded
	assert.NoError(t, validateResumeMetadata(metadata, agents, 1, []string{"s"}, &same))
	rebuilt := recorded
	rebuilt.IndexSHA256 = "b"
	assert.ErrorContains(t, validateResumeMetadata(metadata, agents, 1, []string{"s"}, &rebuilt), "resume retrieval")
	assert.ErrorContains(t, validateResumeMetadata(metadata, agents, 1, []string{"s"}, nil), "resume retrieval")
}
