# Failure causes after retrieving the source page

Every `R1` and `R2` attempt whose searches returned the scenario's source page
but that did not fully succeed (`source_rank > 0`, `full_success: false` in
`attempts.jsonl`). One rater classified each attempt from its search results
and `kubectl` changes. Attempts are named `scenario/attempt`.

| Cause                                                                      | `R1`                                                                                                                                                                | `R2`                                                                                                                                                                                                                                                                                              |
| -------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------- | ------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------------- |
| Wrong diagnosis; the query followed the wrong hypothesis                   | `automount-token-disabled/3`, `daemonset-missing-toleration/1`, `daemonset-missing-toleration/2`, `daemonset-missing-toleration/3`, `service-targetport-mismatch/1` | `automount-token-disabled/1`, `daemonset-missing-toleration/1`, `daemonset-missing-toleration/2`, `daemonset-missing-toleration/3`, `service-targetport-mismatch/3`                                                                                                                               |
| Right diagnosis; the change was rejected                                   | `nonroot-securitycontext/3`, `readonly-rootfs/3`, `resourcequota-blocked/2`, `networkpolicy-cross-namespace/3`                                                      | `statefulset-volumeclaim/3`, `startup-probe-slow-app/3`, `configmap-update-needs-reload/1`                                                                                                                                                                                                        |
| Right diagnosis; the needed fact was in the retrieved text but not applied | `readonly-rootfs/1`, `networkpolicy-egress-dns/3`                                                                                                                   | `networkpolicy-egress-dns/1`, `networkpolicy-egress-dns/2`, `networkpolicy-egress-dns/3`, `stale-image-ifnotpresent/1`                                                                                                                                                                            |
| Right diagnosis; wrong or partial fix the retrieved text does not decide   | `networkpolicy-cross-namespace/1`, `networkpolicy-egress-dns/2`, `overprivileged-serviceaccount/3`, `configmap-missing-key/2`                                       | `overprivileged-serviceaccount/1`, `overprivileged-serviceaccount/2`, `overprivileged-serviceaccount/3`, `resourcequota-blocked/1`, `unschedulable-cpu-request/2`, `immutable-configmap/2`, `immutable-configmap/3`, `rollback-bad-release/2`, `rollback-bad-release/3`, `deprecated-api-group/3` |

Notes per attempt:

- `daemonset-missing-toleration`: queries asked about replicas or
  `desiredNumberScheduled`; no attempt listed the node taints.
- `networkpolicy-egress-dns`: the network-policies page caution "A default
  deny-all egress policy also blocks DNS traffic" was in the results of every
  failed attempt except `R1` attempt 2.
- `readonly-rootfs/1` (`R1`): the results described `emptyDir`; the agent
  mounted a `hostPath`.
- `stale-image-ifnotpresent/1` (`R2`): the results described `imagePullPolicy`;
  the agent re-set the same image tag.
- Rejected changes: merge patches that replaced the container list
  (`containers[0].image: Required value`), immutable `volumeClaimTemplates`,
  patches to a Pod instead of its Deployment, annotations set on the Deployment
  instead of the Pod template.
