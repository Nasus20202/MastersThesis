# Skill sources

This file records provenance for the skill condition. It is outside the embedded `skills/` tree and is not exposed to the model.

## Loader boundary

Progressive disclosure is prompt-guided rather than state-enforced. `load_skill` exposes a skill body; `load_reference` validates the owning skill and exact reference filename but does not track whether the parent skill was loaded earlier. This keeps the benchmark harness stateless and small while preserving all tool calls in result evidence.

## Kubernetes and kubectl

Kubernetes knowledge is based on the frozen common corpus from D-017:

- repository: `kubernetes/website`
- revision: `ea639c1d22a60365d07b78692b1b1a2eb866bd15`
- corpus: `content/en/docs/**/*.md`
- license: CC BY 4.0

Primary source areas used when reviewing the skill content:

| Skill content | Frozen documentation area |
| --- | --- |
| API objects and metadata | `content/en/docs/concepts/overview/working-with-objects/` |
| Architecture and nodes | `content/en/docs/concepts/architecture/` |
| Workloads and Pods | `content/en/docs/concepts/workloads/` |
| Images | `content/en/docs/concepts/containers/images.md` |
| Scheduling and resources | `content/en/docs/concepts/scheduling-eviction/`, `content/en/docs/concepts/configuration/` |
| Services and networking | `content/en/docs/concepts/services-networking/` |
| Storage | `content/en/docs/concepts/storage/` |
| ConfigMaps and Secrets | `content/en/docs/concepts/configuration/` |
| RBAC and authorization | `content/en/docs/reference/access-authn-authz/` |
| Security and policy | `content/en/docs/concepts/security/`, `content/en/docs/concepts/policy/` |
| Autoscaling | `content/en/docs/concepts/workloads/autoscaling/` |
| API extensions | `content/en/docs/concepts/extend-kubernetes/` |
| kubectl commands | `content/en/docs/reference/kubectl/` |

The skill files are concise human-authored summaries, not copied documentation. Scenario definitions and grading criteria are not sources for skill content.

## Bash

The Bash skill summarizes stable shell behavior such as quoting, pipelines, exit status, redirection, files, and processes. Its primary external reference is the GNU Bash Reference Manual. Kubernetes benchmark incidents and expected repairs are not used as Bash-skill source material.

## Troubleshooting

The troubleshooting skill is a generic workflow derived from the project's observed procedural weaknesses and the research goal of testing explicit procedural knowledge. It intentionally contains no scenario-specific repair sequence.