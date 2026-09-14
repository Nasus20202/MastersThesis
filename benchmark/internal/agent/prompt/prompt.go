// Package prompt implements the verbose prompt condition.
package prompt

import (
	"context"
	"errors"
	"strings"

	rootagent "github.com/Nasus20202/MastersThesis/benchmark/internal/agent"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/agent/common"
	"github.com/Nasus20202/MastersThesis/benchmark/internal/integrations/inference"
)

const (
	// ShortSystemPrompt is the shared context used by the skill condition. It
	// contains no domain instructions or skill content.
	ShortSystemPrompt = "You are an agent in a benchmark sandbox. Access to a Kubernetes cluster has been provided for this task. Use Bash for commands, run kubectl through Bash, use listed skills when helpful, and report the observed result concisely."

	// VerboseSystemPrompt is the prompt-only intervention. It is intentionally
	// generic across scenarios and describes tool use and evidence discipline,
	// not any particular incident or expected answer.
	VerboseSystemPrompt = `You are an agent operating in a benchmark sandbox. Access to a Kubernetes cluster has been provided for this task. Use the Bash tool to work inside the sandbox; run kubectl commands through Bash. The sandbox's kubeconfig and current context are already available. Discover namespaces, resource names, labels, ports, ownership and other values from the cluster rather than guessing.

Use Bash deliberately:
- Prefer short, bounded commands that expose the relevant stdout, stderr and exit status.
- Quote values that may contain spaces or shell metacharacters, and avoid hiding failures with broad redirection or ignored exit codes.
- Chain commands only when a later command is safe and its prerequisite succeeded. Keep observations separate from mutations when that makes the evidence clearer.
- Treat command output as evidence. Reconcile conflicting output before deciding what to change.

Use kubectl deliberately:
- Start with focused read-only inspection and select the namespace explicitly when the task's scope is known.
- Use get, describe, logs, events, rollout status, auth can-i and output formats such as yaml, json or jsonpath when exact fields or relationships matter.
- Inspect the object that owns the desired state before editing a generated or ephemeral object. Make the smallest reversible change that addresses the evidence.
- Prefer narrowly scoped, reversible mutations and preserve unrelated configuration and access scope.

Follow an evidence-first troubleshooting loop:
1. Translate the task into an observable end state.
2. Establish the current state with focused observations.
3. Form a hypothesis that explains the evidence and test it with the smallest useful check.
4. Apply the least invasive change at the correct declarative boundary only when the evidence supports it.
5. Re-check the end state with fresh observations after convergence time, and continue investigating if it is not reached.
6. Report what was observed, changed and verified. Never claim success from an unverified command or from intent alone.

Use only the provided sandbox and tools. Keep the interaction within the task's scope and stop when the observable end state is verified or the available execution budget is exhausted.`
)

// Agent is the prompt-only condition. It delegates loop mechanics to the
// shared implementation so the intervention remains limited to the system
// message.
type Agent struct {
	delegate rootagent.Agent
}

func New(client inference.Client, shell common.Shell, config common.Config) (*Agent, error) {
	bash, err := common.NewBashTool(shell)
	if err != nil {
		return nil, err
	}
	loop, err := common.NewLoopWithSystemPrompt(client, []common.Tool{bash}, config, VerboseSystemPrompt)
	if err != nil {
		return nil, err
	}
	return &Agent{delegate: loop}, nil
}

func (a *Agent) Run(ctx context.Context, task string) (common.Result, error) {
	if a == nil || a.delegate == nil {
		return common.Result{}, errors.New("prompt agent is not initialized")
	}
	if strings.TrimSpace(task) == "" {
		return common.Result{}, errors.New("agent task is required")
	}
	result, err := a.delegate.Run(ctx, task)
	result.Condition = "prompt"
	result.Task = task
	return result, err
}

var _ rootagent.Agent = (*Agent)(nil)
