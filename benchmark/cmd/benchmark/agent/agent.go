package agent

import (
	"fmt"
	"os"
	"strings"

	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/baseline"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference/llama"
	sandboxintegration "github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/sandbox"
)

func NewBaselineFactory() (rootagent.Factory, error) {
	model := strings.TrimSpace(os.Getenv("LLAMA_MODEL_NAME"))
	if model == "" {
		return nil, fmt.Errorf("LLAMA_MODEL_NAME is required")
	}

	host := strings.TrimSpace(os.Getenv("LLAMA_CLIENT_HOST"))
	if host == "" {
		host = "127.0.0.1"
	}
	port := strings.TrimSpace(os.Getenv("LLAMA_PORT"))
	if port == "" {
		port = "8080"
	}
	client, err := llama.NewClient(llama.Config{
		BaseURL: fmt.Sprintf("http://%s:%s", host, port),
		Model:   model,
	})
	if err != nil {
		return nil, fmt.Errorf("create llama client: %w", err)
	}
	inferenceClient, err := llama.NewAdapter(client)
	if err != nil {
		return nil, fmt.Errorf("create inference adapter: %w", err)
	}
	loopConfig := common.DefaultConfig()
	return func(shell sandboxintegration.Executor) (rootagent.Agent, error) {
		return baseline.New(inferenceClient, shell, loopConfig)
	}, nil
}
