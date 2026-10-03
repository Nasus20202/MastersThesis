package common

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFailedChangeInReadsCommandShapeAndErrorLine(t *testing.T) {
	tests := []struct {
		name     string
		evidence CommandEvidence
		want     FailedChange
		query    string
	}{
		{
			name: "merge patch dropping the image",
			evidence: CommandEvidence{
				Command:  `kubectl patch deployment app -n default --type='merge' -p '{"spec": {"template": {"spec": {"containers": [{"name": "app", "args": ["nginx"]}]}}}}'`,
				Stderr:   `The Deployment "app" is invalid: spec.template.spec.containers[0].image: Required value` + "\n",
				ExitCode: 1,
			},
			want:  FailedChange{Verb: "patch", Kind: "deployment", Error: "The Deployment is invalid: spec.template.spec.containers[0].image: Required value"},
			query: "kubectl patch deployment The Deployment is invalid: spec.template.spec.containers[0].image: Required value",
		},
		{
			name: "server error prefix and flag values",
			evidence: CommandEvidence{
				Command:  `kubectl patch --type merge -p '{"spec": {}}' deployment/app`,
				Stderr:   "Error from server (BadRequest): invalid JSON patch\n",
				ExitCode: 1,
			},
			want:  FailedChange{Verb: "patch", Kind: "deployment", Error: "invalid JSON patch"},
			query: "kubectl patch deployment invalid JSON patch",
		},
		{
			name: "inline manifest after a warning",
			evidence: CommandEvidence{
				Command:  "kubectl get pods\ncat <<EOF | kubectl apply -f -\napiVersion: v1\nkind: Pod\nmetadata:\n  name: app\nEOF",
				Stderr:   "Warning: resource pods/app is missing the last-applied annotation\nThe Pod \"app\" is invalid: spec: Forbidden: pod updates may not change fields\n",
				ExitCode: 1,
			},
			want: FailedChange{Verb: "apply", Kind: "pod", Error: "The Pod is invalid: spec: Forbidden: pod updates may not change fields"},
		},
		{
			name: "set subcommand",
			evidence: CommandEvidence{
				Command:  "kubectl set image deployment/app app=nginx:2 && kubectl rollout status deployment/app",
				Stderr:   "error: unable to find container named \"web\"\n",
				ExitCode: 1,
			},
			want: FailedChange{Verb: "set", Kind: "deployment", Error: "unable to find container named"},
		},
		{
			name: "kind from the rejection",
			evidence: CommandEvidence{
				Command:  "kubectl apply -f fixed.yaml",
				Stderr:   `The StatefulSet "app" is invalid: spec: Forbidden: updates to statefulset spec are forbidden`,
				ExitCode: 1,
			},
			want: FailedChange{Verb: "apply", Kind: "statefulset", Error: "The StatefulSet is invalid: spec: Forbidden: updates to statefulset spec are forbidden"},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			change, ok := FailedChangeIn(test.evidence)
			assert.True(t, ok)
			assert.Equal(t, test.want, change)
			if test.query != "" {
				assert.Equal(t, test.query, change.Query())
			}
		})
	}
}

func TestFailedChangeInIgnoresReadsSuccessesTimeoutsAndWarnings(t *testing.T) {
	for name, evidence := range map[string]CommandEvidence{
		"read":                                  {Command: "kubectl get deployment app", Stderr: "Error from server (NotFound): not found", ExitCode: 1},
		"success":                               {Command: "kubectl patch deployment app -p '{}'", Stdout: "deployment.apps/app patched", ExitCode: 0},
		"timeout":                               {Command: "kubectl delete pod app", Stderr: "", ExitCode: -1},
		"dry run":                               {Command: "kubectl apply --dry-run=server -f app.yaml", Stderr: "error: invalid", ExitCode: 1},
		"only a warn":                           {Command: "kubectl delete pod app --force", Stderr: "Warning: Immediate deletion does not wait\n", ExitCode: 1},
		"write succeeded, later command failed": {Command: "kubectl apply -f app.yaml && grep ready status.txt", Stdout: "deployment.apps/app created\n", ExitCode: 1},
	} {
		t.Run(name, func(t *testing.T) {
			_, ok := FailedChangeIn(evidence)
			assert.False(t, ok)
		})
	}
}

func TestFailedChangeInReadsRedirectedErrorFromStdout(t *testing.T) {
	change, ok := FailedChangeIn(CommandEvidence{Command: "kubectl patch deployment app -p '{' 2>&1", Stdout: "Error from server (BadRequest): invalid JSON patch\n", ExitCode: 1})
	assert.True(t, ok)
	assert.Equal(t, "invalid JSON patch", change.Error)
}

func TestFailedChangeInCleansMultilineAndEchoedObjects(t *testing.T) {
	change, ok := FailedChangeIn(CommandEvidence{
		Command:  "kubectl patch service app -p '{}'",
		Stderr:   "The Service \"app\" is invalid: \n* spec.ports[0].targetPort: Invalid value: \"80/TCP\": must contain only alpha-numeric characters\n* other\n",
		ExitCode: 1,
	})
	assert.True(t, ok)
	assert.Equal(t, "The Service is invalid: spec.ports[0].targetPort: Invalid value: must contain only alpha-numeric characters", change.Error)

	change, ok = FailedChangeIn(CommandEvidence{
		Command:  "kubectl patch statefulset app -p '{}'",
		Stderr:   `The StatefulSet "app" is invalid: spec.volumeClaimTemplates: Invalid value: [{"a":"b","c":{"d":null,"e":30,"f":false}}]: field is immutable`,
		ExitCode: 1,
	})
	assert.True(t, ok)
	assert.Equal(t, "The StatefulSet is invalid: spec.volumeClaimTemplates: Invalid value: field is immutable", change.Error)

	_, ok = FailedChangeIn(CommandEvidence{Command: "kubectl patch pod app -p '{", Stderr: "bash: -c: line 1: unexpected EOF while looking for matching `''\n", ExitCode: 2})
	assert.False(t, ok)
}
