# Development scenario catalogue

Evaluator-only. This file records scenario states, expected repairs and grading.
It must never be exposed to a model, used as retrieval context, or included in
fine-tuning data. The runner exposes only the `task` field from each
`benchmark/scenarios/<id>/scenario.yaml`.

The contract, lifecycle, visibility rules and scoring are in
[benchmark.md](benchmark.md). Open decisions are recorded only in the
[decision log](decision-log.md).

## Source corpus

Scenarios are grounded in the frozen corpus `kubernetes/website@ea639c1d22a60365d07b78692b1b1a2eb866bd15`
(see [sources.md](sources.md)). Each scenario gets a `source.md` with its exact
corpus path, blob SHA and section.

## Difficulty

| Band       | Design profile                                                             | Expected baseline full success |
| ---------- | -------------------------------------------------------------------------- | ------------------------------ |
| **Easy**   | one obvious defect, single-field fix, one subsystem                        | ≥ 0.70                         |
| **Medium** | one defect needing inference across objects, or a constrained two-step fix | 0.30 – 0.70                    |
| **Hard**   | coupled defects, cross-subsystem repair, or several independent parts      | < 0.30 (floor-safe, > 0)       |
| **Anchor** | trivially solved; sanity/regression check only                             | ~1.00                          |

Target over graded scenarios: **25% easy / 50% medium / 25% hard**; anchors are
excluded from the balance and capped at two. Labels are design hypotheses
calibrated by the baseline run. Difficulty is evaluator metadata, never shown to
the model.

Current tally: **11 easy / 23 medium / 11 hard** across 45 graded scenarios (24 / 51 / 24),
plus 2 anchors.

## Cluster profiles and add-ons

Scenarios run in fresh disposable single-node `kind` clusters. Each declares one profile:

- **default** — kindnet, most scenarios.
- **calico** — `disableDefaultCNI` + pinned Calico (`podSubnet: 192.168.0.0/16`), for scenarios that need NetworkPolicy enforcement. Installed and waited on in `prepare`.
- **registry** — default kindnet plus an in-cluster registry mirror; the registry is deployed by the scenario in `prepare`.

Every profile also includes:

- shared **pull-through registry caches** (`registry:3.1.1`) on the kind network, mirrored for `docker.io` and `quay.io` in `containerdConfigPatches`, so pinned images are pulled once and reused across clusters;
- kind's built-in default StorageClass `standard` (`rancher.io/local-path`), no add-on; its `WaitForFirstConsumer` binding means a PVC binds only after a consumer Pod is scheduled.

Per-scenario add-ons, installed evaluator-side in `prepare`, never part of the task:

- **Traefik** for the three ingress/gateway scenarios;
- an in-cluster registry for the image scenarios (`private-registry-pull-secret` adds basic authentication; `stale-image-ifnotpresent` uses a mutable tag).

All images, manifests and binaries are pinned and served from the local cache.

## Scenario index

Type: T troubleshooting, C constructive. Difficulty: E easy, M medium, H hard,
A anchor. Knowledge: G general, D documentation-dependent, M multi-source,
V version-specific.

| Scenario                           | Area                   | Type | Diff | Know | Env      |
| ---------------------------------- | ---------------------- | ---- | ---- | ---- | -------- |
| `deployment-selector-mismatch`     | workloads              | T    | M    | G    | default  |
| `rollback-bad-release`             | workloads              | T    | M    | D    | default  |
| `daemonset-missing-toleration`     | scheduling             | T    | M    | D    | default  |
| `job-backoff-exhausted`            | workloads              | T    | E    | G    | default  |
| `statefulset-headless-service`     | workloads + networking | T    | H    | M    | default  |
| `readiness-probe-breaks-endpoints` | pods + networking      | T    | M    | D    | default  |
| `liveness-restart-loop`            | pods                   | T    | M    | D    | default  |
| `startup-probe-slow-app`           | pods                   | C    | E    | D    | default  |
| `stuck-terminating-pod`            | lifecycle              | T    | M    | G    | default  |
| `sidecar-shared-volume`            | pods + configuration   | C    | H    | M    | default  |
| `container-crash-loop`             | workloads + pods       | T    | H    | G    | default  |
| `image-pull-failure`               | images                 | T    | A    | G    | default  |
| `stale-image-ifnotpresent`         | images                 | T    | M    | D    | registry |
| `private-registry-pull-secret`     | images                 | C    | E    | D    | registry |
| `unschedulable-cpu-request`        | resources              | T    | E    | G    | default  |
| `node-selector-no-match`           | scheduling             | T    | E    | G    | default  |
| `taint-toleration-placement`       | scheduling             | C    | E    | D    | default  |
| `antiaffinity-unschedulable`       | scheduling             | T    | M    | D    | default  |
| `oom-memory-limit`                 | resources              | T    | E    | G    | default  |
| `qos-guaranteed-requests`          | resources              | C    | E    | D    | default  |
| `service-targetport-mismatch`      | networking             | T    | M    | D    | default  |
| `headless-service-dns`             | networking             | C    | M    | D    | default  |
| `dns-name-resolution`              | networking             | T    | M    | M    | default  |
| `ingress-path-routing`             | networking             | C    | M    | D    | traefik  |
| `gateway-httproute-basic`          | networking             | C    | M    | D    | traefik  |
| `gateway-route-misconfig`          | networking             | T    | H    | D    | traefik  |
| `networkpolicy-blocked`            | networking             | T    | H    | D    | calico   |
| `networkpolicy-egress-dns`         | networking             | T    | H    | M    | calico   |
| `networkpolicy-cross-namespace`    | networking             | T    | M    | D    | calico   |
| `pvc-pending`                      | storage                | T    | M    | D    | default  |
| `bind-precreated-pv`               | storage                | C    | M    | D    | default  |
| `statefulset-volumeclaim`          | storage + workloads    | T    | H    | M    | default  |
| `hostpath-fsgroup`                 | storage + security     | T    | M    | M    | default  |
| `configmap-missing-key`            | configuration          | T    | E    | G    | default  |
| `secret-ref-mismatch`              | configuration          | T    | E    | G    | default  |
| `immutable-configmap`              | configuration          | T    | M    | V    | default  |
| `configmap-update-needs-reload`    | configuration          | T    | H    | D    | default  |
| `missing-rbac-binding`             | authorization          | T    | H    | G    | default  |
| `overprivileged-serviceaccount`    | authorization          | C    | H    | D    | default  |
| `automount-token-disabled`         | authorization          | T    | M    | D    | default  |
| `nonroot-securitycontext`          | security               | T    | M    | D    | default  |
| `readonly-rootfs`                  | security               | T    | M    | D    | default  |
| `psa-restricted-namespace`         | security               | T    | H    | V    | default  |
| `resourcequota-blocked`            | governance             | T    | M    | D    | default  |
| `limitrange-defaults`              | governance             | C    | E    | D    | default  |
| `deprecated-api-group`             | api                    | C    | M    | V    | default  |
| `basic-deploy-service`             | workloads + networking | C    | A    | G    | default  |

## Scenario detail

### Workload controllers

- **`deployment-selector-mismatch`** — **Fault:** `spec.selector.app=web` but template `app=app`; no ReplicaSet created. **Fix:** replace the Deployment with a matching selector (selector is immutable). **Grade:** available 3 [2]; selector matches template and RS owned [1]. **Note:** may stall baseline on the immutable rule.
- **`rollback-bad-release`** — **Fault:** rev1 healthy, rev2 bad, available<3. **Fix:** restore rev1 via `rollout undo` or re-apply the good image. **Grade:** available [2]; running image == last good [1]; history preserved [1]. **Note:** requires rollout-history reasoning, not Pod deletion.
- **`daemonset-missing-toleration`** — **Fault:** DS desired<N; node carries a custom `NoSchedule` taint. **Fix:** add the matching toleration, keep the taint. **Grade:** `numberReady==desiredNumberScheduled` [2]; taint intact [1].
- **`job-backoff-exhausted`** — **Fault:** Job hits `BackoffLimitExceeded`, 0 succeeded. **Fix:** correct the command/args so the Job completes. **Grade:** `succeeded>=1` [2]; no CrashLoop/Error Pod [1].
- **`statefulset-headless-service`** — **Fault:** StatefulSet Pods never Ready; the `serviceName` headless Service is absent/mismatched. **Fix:** create/fix the headless Service (`clusterIP: None`, selector/ports). **Grade:** Pods Ready [2]; headless Service selects Pods [2]; per-Pod DNS resolves [1]. **Note:** links the StatefulSet contract to the missing Service.

### Pods, probes and lifecycle

- **`readiness-probe-breaks-endpoints`** — **Fault:** Pods Running but never Ready (bad readiness path/port); no ready endpoints. **Fix:** correct the readiness probe. **Grade:** Pods Ready [2]; endpoints ready [2]; service responds [1]. **Note:** Pods look healthy, so trace the symptom to endpoints.
- **`liveness-restart-loop`** — **Fault:** over-eager liveness kills a healthy container. **Fix:** correct the liveness probe. **Grade:** Pod Ready [2]; restart count stable over a window [2].
- **`startup-probe-slow-app`** — **Fault:** slow app killed by liveness before serving. **Fix:** add a startup probe and correct liveness/readiness. **Grade:** Pod Ready [2]; no restarts in the window [1].
- **`stuck-terminating-pod`** — **Fault:** Pod stuck Terminating (finalizer, ignores TERM), Deployment below replicas. **Fix:** remove the finalizer or force-delete; ensure a healthy replacement. **Grade:** no Terminating Pod beyond the bound [2]; Deployment available [1].
- **`sidecar-shared-volume`** — **Fault:** main container serves an empty directory. **Fix:** add a sidecar writing into a shared `emptyDir` mounted by both. **Grade:** both containers Ready [1]; emptyDir mounted by both [1]; content served [2].
- **`container-crash-loop`** — **Fault:** invalid container args cause CrashLoopBackOff. **Fix:** fix the Deployment pod template; deleting Pods must not count. **Grade:** Deployment ready [2]; 3 ready replicas [1]. **Note:** baseline often edits the wrong object.

### Images

- **`image-pull-failure`** (anchor) — **Fault:** unpullable image → `ImagePullBackOff`. **Fix:** correct the image reference. **Grade:** Deployment ready [2]; all replicas on the correct image [1].
- **`stale-image-ifnotpresent`** — **Fault:** mutable tag republished; `IfNotPresent` keeps stale layers. **Fix:** force a fresh pull (`Always`) or move to a digest. **Grade:** running image == expected digest [2]; Deployment ready [1].
- **`private-registry-pull-secret`** — **Fault:** requires-auth image with no `imagePullSecrets` → `ImagePullBackOff`. **Fix:** create a `kubernetes.io/dockerconfigjson` Secret and reference it. **Grade:** Secret type/reference correct [1]; Pod Ready [2]. **Note:** heaviest scenario; first to move to spares if flaky.

### Scheduling, nodes and resources

- **`unschedulable-cpu-request`** — **Fault:** Pending, CPU request exceeds allocatable. **Fix:** reduce or remove the request. **Grade:** Deployment ready [2]; 3 ready replicas [1]. **Note:** add a criterion to block the `replicas=1` workaround.
- **`node-selector-no-match`** — **Fault:** Pending `FailedScheduling`; `nodeSelector` matches no node. **Fix:** remove or correct the `nodeSelector`. **Grade:** all replicas Running/Ready [2].
- **`taint-toleration-placement`** — **Fault:** node tainted, workload has no toleration. **Fix:** add a matching toleration. **Grade:** Pods on the tainted node [2]; taint intact [1].
- **`antiaffinity-unschedulable`** — **Fault:** required anti-affinity unsatisfiable on one node. **Fix:** relax to preferred or restructure labels. **Grade:** all replicas Ready [2]; no impossible required rule [1].
- **`oom-memory-limit`** — **Fault:** memory limit below working set → OOMKilled restart loop. **Fix:** raise the limit and set a matching request. **Grade:** Pod Ready [2]; restarts stable [1]. **Note:** needs a pinned allocator image; verify deterministic OOM.
- **`qos-guaranteed-requests`** — **Fault:** workload has no resources; QoS BestEffort. **Fix:** set requests equal to limits on every container. **Grade:** `qosClass==Guaranteed` [2]; Deployment ready [1].

### Networking

- **`service-targetport-mismatch`** — **Fault:** selector matches and endpoints exist, but `targetPort` is wrong. **Fix:** correct `targetPort` (numeric or named). **Grade:** service responds [2]; endpoints ready [1]; Pods Ready [1].
- **`headless-service-dns`** — **Fault:** StatefulSet needs stable per-Pod DNS; no Service exists. **Fix:** create a headless Service with correct selector/ports. **Grade:** headless Service correct [2]; per-Pod names resolve to Pod IPs [2].
- **`dns-name-resolution`** — **Fault:** app cannot resolve a peer Service (wrong name/qualifier). **Fix:** correct the client config to a resolvable name. **Grade:** resolution succeeds [2]; app ready [1].
- **`ingress-path-routing`** — **Fault:** Service exists; must expose it via Ingress on a path. **Fix:** create an Ingress with correct class, host/path and backend. **Grade:** Ingress admitted [1]; path route works [2].
- **`gateway-httproute-basic`** — **Fault:** Service exists; must expose it via Gateway API. **Fix:** create a Gateway and an HTTPRoute to the Service. **Grade:** Gateway programmed [1]; route accepted [1]; request succeeds [2]. **Note:** corpus page is conceptual; field knowledge is external.
- **`gateway-route-misconfig`** — **Fault:** Gateway programmed but HTTPRoute does not attach (parentRef/sectionName/ReferenceGrant). **Fix:** fix route attachment and any ReferenceGrant. **Grade:** route accepted [1]; request through gateway succeeds [2].
- **`networkpolicy-blocked`** — **Fault:** default-deny NetworkPolicy blocks required traffic. **Fix:** add a policy allowing only the intended peer/port. **Grade:** allowed path works [2]; disallowed path still blocked [2]. **Note:** the negative criterion blocks "delete all".
- **`networkpolicy-egress-dns`** — **Fault:** default-deny egress blocks name resolution; app cannot reach its dependency. **Fix:** allow egress to kube-system DNS (UDP/TCP 53) and the required peer, keeping other egress denied. **Grade:** dependency reachable [2]; DNS resolution works [2]; disallowed egress still blocked [1]. **Note:** forgetting DNS is the classic trap.
- **`networkpolicy-cross-namespace`** — **Fault:** policy matches the client label but not its namespace, so traffic is blocked. **Fix:** add the correct `namespaceSelector`. **Grade:** allowed client reaches the app [2]; other namespaces still blocked [2].

### Storage

- **`pvc-pending`** — **Fault:** PVC Pending (bad class/access mode/size). **Fix:** correct the claim or provide the class. **Grade:** PVC bound [2]; Pod Ready [1].
- **`bind-precreated-pv`** — **Fault:** a static PersistentVolume exists to be consumed. **Fix:** create a matching PVC (empty class, selector/capacity). **Grade:** bound to the intended PV [2]; Pod Ready [1].
- **`statefulset-volumeclaim`** — **Fault:** `volumeClaimTemplates` cannot bind; Pods never start. **Fix:** provide/correct the StorageClass or claim template. **Grade:** PVCs bound [2]; StatefulSet Pods Ready [2].
- **`hostpath-fsgroup`** — **Fault:** non-root container denied the volume by ownership. **Fix:** set pod `fsGroup` and/or an init chown. **Grade:** volume usable [2]; Pod Ready [1].

### Configuration

- **`configmap-missing-key`** — **Fault:** `CreateContainerConfigError`: referenced key absent. **Fix:** add the key or correct the reference. **Grade:** Pod Ready [2]; key consumed [1].
- **`secret-ref-mismatch`** — **Fault:** template references a missing Secret name/key. **Fix:** create/correct the Secret or fix the reference. **Grade:** Pod Ready [2]; Secret present [1].
- **`immutable-configmap`** — **Fault:** update rejected because `immutable: true`. **Fix:** recreate the ConfigMap and restart consumers. **Grade:** value updated [2]; Pod Ready consuming new value [1].
- **`configmap-update-needs-reload`** — **Fault:** ConfigMap updated but Pods use the old value (env/subPath). **Fix:** trigger a rollout. **Grade:** running config matches [2]; Deployment ready [1].

### Authorization, security and policy

- **`missing-rbac-binding`** — **Fault:** init container gets 403; no Role/RoleBinding. **Fix:** create a scoped Role and RoleBinding. **Grade:** Deployment ready (init succeeds) [2]; least-privilege [2]. **Note:** over-granting fails.
- **`overprivileged-serviceaccount`** — **Fault:** app works but its Role grants unused resources/verbs. **Fix:** reduce to the exact resource and verbs used. **Grade:** app still works [2]; excess rules removed [2].
- **`automount-token-disabled`** — **Fault:** no API token mounted. **Fix:** enable automount or use a projected token; ensure RBAC allows the call. **Grade:** API call succeeds [2]; Pod Ready [1].
- **`nonroot-securitycontext`** — **Fault:** `runAsNonRoot` conflicts with an image whose user is root. **Fix:** set `runAsUser`/`fsGroup` consistent with the image. **Grade:** Pod Ready [2]; effective UID non-zero [1].
- **`readonly-rootfs`** — **Fault:** read-only root filesystem breaks an app writing temp files. **Fix:** add a writable `emptyDir` at the needed path. **Grade:** app works [2]; root filesystem still read-only [1].
- **`psa-restricted-namespace`** — **Fault:** Pod rejected by Pod Security Admission (`restricted`). **Fix:** make the Pod compliant, not the namespace laxer. **Grade:** Pod admitted [2]; namespace policy intact [1].
- **`resourcequota-blocked`** — **Fault:** ResourceQuota rejects new Pods. **Fix:** fit the workload within quota or raise it intentionally. **Grade:** Pods admitted [2]; quota respected [1].
- **`limitrange-defaults`** — **Fault:** namespace requires default requests/limits. **Fix:** create a LimitRange; confirm a new Pod adopts defaults. **Grade:** LimitRange present [1]; defaults applied [2].

### API

- **`deprecated-api-group`** — **Fault:** manifest uses a removed group/version. **Fix:** produce an equivalent resource in the served version. **Grade:** created under the correct group/version [2]; behaviour present [1].

### Anchors

- **`image-pull-failure`** — defined under Images.
- **`basic-deploy-service`** — **Fault:** empty namespace. **Fix:** deploy a workload and expose it with a ClusterIP Service. **Grade:** Deployment ready [1]; service responds [1]. **Note:** sanity check for the constructive harness.
