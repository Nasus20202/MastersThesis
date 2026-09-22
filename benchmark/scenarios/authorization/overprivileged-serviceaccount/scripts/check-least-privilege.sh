#!/bin/sh
set -eu

rule_count="$(kubectl get role api-reader -o jsonpath='{range .rules[*]}x{end}')"
api_group_count="$(kubectl get role api-reader -o jsonpath='{range .rules[0].apiGroups[*]}x{end}')"
resource_count="$(kubectl get role api-reader -o jsonpath='{range .rules[0].resources[*]}x{end}')"
verb_count="$(kubectl get role api-reader -o jsonpath='{range .rules[0].verbs[*]}x{end}')"
api_group="$(kubectl get role api-reader -o jsonpath='{.rules[0].apiGroups[0]}')"
resource="$(kubectl get role api-reader -o jsonpath='{.rules[0].resources[0]}')"
verb="$(kubectl get role api-reader -o jsonpath='{.rules[0].verbs[0]}')"

printf 'rules=%s api_group=%s resource=%s verb=%s\n' "$rule_count" "$api_group" "$resource" "$verb"

test "$rule_count" = "x"
test "$api_group_count" = "x"
test "$resource_count" = "x"
test "$verb_count" = "x"
test "$api_group" = ""
test "$resource" = "configmaps"
test "$verb" = "get"
