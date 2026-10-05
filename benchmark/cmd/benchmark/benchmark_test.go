package main

import (
	"testing"

	"github.com/stretchr/testify/assert"

	commandagent "github.com/Nasus20202/MastersThesis/benchmark/cmd/benchmark/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/results"
)

func TestValidateResumeMetadataComparesRetrieval(t *testing.T) {
	recorded := results.RetrievalProvenance{Mode: "hybrid", Chunking: "windows", TopK: 5, MaxBytes: 8192, IndexSHA256: "a"}
	metadata := results.RunMetadata{RunType: results.RunTypeBenchmark, Agents: []string{"rag"}, RepeatCount: 1, Scenarios: []string{"s"}, Retrieval: &recorded}
	agents := []commandagent.Name{commandagent.RAG}

	same := recorded
	assert.NoError(t, validateResumeMetadata(metadata, agents, 1, []string{"s"}, &same, nil))
	rebuilt := recorded
	rebuilt.IndexSHA256 = "b"
	assert.ErrorContains(t, validateResumeMetadata(metadata, agents, 1, []string{"s"}, &rebuilt, nil), "resume retrieval")
	assert.ErrorContains(t, validateResumeMetadata(metadata, agents, 1, []string{"s"}, nil, nil), "resume retrieval")
}

func TestValidateResumeMetadataComparesSampling(t *testing.T) {
	recorded := results.SamplingProvenance{Model: "gemma", Sampling: inference.Sampling{Temperature: 1, TopK: 64, TopP: 0.95}}
	metadata := results.RunMetadata{RunType: results.RunTypeBenchmark, Agents: []string{"prompt"}, RepeatCount: 1, Scenarios: []string{"s"}, Sampling: &recorded}
	agents := []commandagent.Name{commandagent.Prompt}

	same := recorded
	assert.NoError(t, validateResumeMetadata(metadata, agents, 1, []string{"s"}, nil, &same))
	changed := recorded
	changed.TopK = 40
	assert.ErrorContains(t, validateResumeMetadata(metadata, agents, 1, []string{"s"}, nil, &changed), "resume sampling")
}
