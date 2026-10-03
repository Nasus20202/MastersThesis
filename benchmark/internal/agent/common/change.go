package common

import (
	"context"
	"regexp"
	"slices"
	"strings"
)

// FailedChange is a kubectl command that writes to the cluster and whose Bash
// call exited non-zero, as the command text and its error output show.
type FailedChange struct {
	Verb  string
	Kind  string
	Error string
}

// Query is the documentation search for the failure: the command shape and
// the error line, without object names.
func (c FailedChange) Query() string {
	return strings.Join(slices.DeleteFunc([]string{"kubectl", c.Verb, c.Kind, c.Error}, func(part string) bool { return part == "" }), " ")
}

// Hint returns text to append to the model-visible result of a failed change
// and evidence to record with it. Empty text appends nothing. A Hint is called
// sequentially within one attempt.
type Hint func(context.Context, FailedChange) (text string, details any)

// changeVerbs are kubectl subcommands that write to the cluster. The run
// analysis (internal/analysis) applies the same heuristic to recorded runs.
var changeVerbs = map[string]bool{
	"apply": true, "patch": true, "edit": true, "set": true, "delete": true, "create": true,
	"replace": true, "scale": true, "label": true, "annotate": true, "taint": true,
	"cordon": true, "uncordon": true, "drain": true, "expose": true, "autoscale": true,
}

// flagsWithValue are kubectl flags whose value is a separate word.
var flagsWithValue = map[string]bool{
	"-n": true, "--namespace": true, "--context": true, "--kubeconfig": true,
	"-f": true, "--filename": true, "-p": true, "--patch": true, "--patch-file": true,
	"--type": true, "-l": true, "--selector": true, "-o": true, "--output": true, "-c": true, "--container": true,
}

var (
	commandSeparator = regexp.MustCompile(`\|\||&&|[|;&\n]`)
	manifestKind     = regexp.MustCompile(`(?m)^\s*kind:\s*([A-Za-z]+)`)
	invalidKind      = regexp.MustCompile(`^The ([A-Za-z]+) "`)
	quoted           = regexp.MustCompile(`"[^"]*"`)
	shellQuoted      = regexp.MustCompile(`'[^']*'|"[^"]*"`)
	stdoutError      = regexp.MustCompile(`^(?:Error from server|error:|The [A-Za-z]+ ".*" is invalid)`)
	errorPrefix      = regexp.MustCompile(`^(Error from server \([A-Za-z]+\): |error: )`)
	// objectDump is what remains of a rejected object echoed in the error
	// once its quoted strings are removed, such as {:{:,:null,:30}}; the
	// bracket must be followed by punctuation so that containers[0] stays.
	objectDump = regexp.MustCompile(`[\[{][:,\[\]{} ]+(?:(?:null|true|false|-?\d+)[:,\[\]{} ]*)*`)
	emptyValue = regexp.MustCompile(`:(?:\s*:)+`)
)

// maxErrorBytes bounds the error part of a query; long errors repeat the
// rejected object.
const maxErrorBytes = 200

// FailedChangeIn reports the first kubectl write in a failed Bash call. It is a
// text heuristic: a non-zero exit, a kubectl write in the command and a
// non-warning error line in the output. A timeout (negative exit) is not a
// failed change.
func FailedChangeIn(evidence CommandEvidence) (FailedChange, bool) {
	if evidence.ExitCode <= 0 {
		return FailedChange{}, false
	}
	args := firstWrite(evidence.Command)
	if args == nil {
		return FailedChange{}, false
	}
	line := errorLine(evidence.Stderr)
	if line == "" {
		// kubectl reports errors on stderr; stdout counts only when the agent
		// redirected it and the line reads as an error, not as a success
		// message of a write followed by another failing command.
		if line = errorLine(evidence.Stdout); !stdoutError.MatchString(line) {
			return FailedChange{}, false
		}
	}
	change := FailedChange{Verb: args[0], Kind: commandKind(args, evidence.Command)}
	if change.Kind == "" {
		if match := invalidKind.FindStringSubmatch(line); match != nil {
			change.Kind = strings.ToLower(match[1])
		}
	}
	line = errorPrefix.ReplaceAllString(line, "")
	line = objectDump.ReplaceAllString(quoted.ReplaceAllString(line, ""), " ")
	line = emptyValue.ReplaceAllString(line, ":")
	line = strings.Join(strings.Fields(line), " ")
	if len(line) > maxErrorBytes {
		line = strings.TrimSpace(line[:maxErrorBytes])
	}
	change.Error = line
	return change, true
}

// firstWrite returns the positional arguments of the first kubectl write in
// a command. Quoted strings, such as patch bodies, become one placeholder word
// so their spaces and separators do not split the command.
func firstWrite(command string) []string {
	command = shellQuoted.ReplaceAllString(command, "''")
	for _, segment := range commandSeparator.Split(command, -1) {
		if strings.Contains(segment, "--dry-run") {
			continue
		}
		words := strings.Fields(segment)
		start := slices.Index(words, "kubectl")
		if start < 0 {
			continue
		}
		var args []string
		for index := start + 1; index < len(words); index++ {
			word := words[index]
			if strings.HasPrefix(word, "-") {
				if flagsWithValue[word] {
					index++
				}
				continue
			}
			args = append(args, word)
		}
		if len(args) > 0 && (changeVerbs[args[0]] || args[0] == "rollout" && len(args) > 1 && (args[1] == "restart" || args[1] == "undo")) {
			return args
		}
	}
	return nil
}

// commandKind is the resource type named in the command, such as deployment
// for "patch deployment/app" or "set image deployment/app", or the kind of an
// inline manifest.
func commandKind(args []string, command string) string {
	position := 1
	if args[0] == "set" || args[0] == "rollout" {
		position = 2
	}
	if position < len(args) {
		kind, _, _ := strings.Cut(args[position], "/")
		kind, _, _ = strings.Cut(kind, ".")
		if kind != "" && !strings.ContainsAny(kind, `'"{[$<`+"`") {
			return strings.ToLower(kind)
		}
	}
	if match := manifestKind.FindStringSubmatch(command); match != nil {
		return strings.ToLower(match[1])
	}
	return ""
}

// errorLine is the first non-empty output line that is neither a warning nor
// a shell error, joined with the next line when it ends in a colon, as in
// "The Service "app" is invalid:" followed by "* spec.ports[0]...".
func errorLine(output string) string {
	var lines []string
	for line := range strings.SplitSeq(output, "\n") {
		if line = strings.TrimSpace(line); line != "" {
			lines = append(lines, line)
		}
	}
	for index, line := range lines {
		if strings.HasPrefix(line, "Warning:") || strings.HasPrefix(line, "bash:") {
			continue
		}
		if strings.HasSuffix(line, ":") && index+1 < len(lines) {
			line += " " + strings.TrimPrefix(lines[index+1], "* ")
		}
		return line
	}
	return ""
}
