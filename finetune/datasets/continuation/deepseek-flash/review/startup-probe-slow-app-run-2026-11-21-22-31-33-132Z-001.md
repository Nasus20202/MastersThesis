---
attempt: finetune/datasets/continuation/deepseek-flash/runs/run-2026-11-21-22-31-33-132Z/startup-probe-slow-app/prompt/001.json
verdict: accepted
reason: 'Raises both probe initial delays to 30s, verified by 1/1 Available, delay=30s, and a Ready pod with 0 restarts.'
---
# startup-probe-slow-app · deepseek-flash · run-2026-11-21-22-31-33-132Z · attempt 1

## Task

A slow-starting application Pod is killed before it can serve requests and never becomes Ready. Restore it to a stable, Ready state.

Documented in:
- content/en/docs/tasks/configure-pod-container/configure-liveness-readiness-startup-probes.md: Protect slow starting containers with startup probes

## Inherited stop

The first 53 messages come from the Gemma attempt; it stopped on:

[tool] exit_code: 0
stdout:
Name:             app-68bc7878f-4sc2q
Namespace:        default
Priority:         0
Service Account:  default
Node:             benchmark-startup-probe-slow-app-e13361a6-control-plane/172.19.0.7
Start Time:       Sat, 10 Oct 2026 01:20:51 +0000
Labels:           app=app
                  pod-template-hash=68bc7878f
Annotations:      <none>
Status:           Running
IP:               10.244.0.7
IPs:
  IP:           10.244.0.7
Controlled By:  ReplicaSet/app-68bc7878f
Containers:
  app:
    Container ID:  containerd://c8c4c8501da87a789e4f7c02fc2d3f84c6afae07e2c9738848684eab118cbf40
    Image:         nginx:1.31.5
    Image ID:      docker.io/library/nginx@sha256:05b8cb60c354a44ab824ea6e7dc69b46d50762cdbe728a347a5b656e6fb3d7c4
    Port:          80/TCP (http)
    Host Port:     0/TCP (http)
    Command:
      sh
      -c
      sleep 20; exec nginx -g 'daemon off;'
    State:          Waiting
      Reason:       CrashLoopBackOff
    Last State:     Terminated
      Reason:       Error
      Exit Code:    137
      Started:      Sat, 10 Oct 2026 01:24:08 +0000
      Finished:     Sat, 10 Oct 2026 01:24:12 +0000
    Ready:          False
    Restart Count:  6
    Liveness:       http-get http://:80/ delay=1s timeout=1s period=1s successThreshold=1 failureThreshold=1
    Readiness:      http-get http://:80/ delay=0s timeout=1s period=5s successThreshold=1 failureThreshold=3
    Environment:    <none>
    Mounts:
      /var/run/secrets/kubernetes.io/serviceaccount from kube-api-access-9pg4z (ro)
Conditions:
  Type                        Status
  PodReadyToStartContainers   True 
  Initialized                 True 
  Ready                       False 
  ContainersReady             False 
  PodScheduled                True 
Volumes:
  kube-api-access-9pg4z:
    Type:                    Projected (a volume that contains injected data from multiple sources)
    TokenExpirationSeconds:  3607
    ConfigMapName:           kube-root-ca.crt
    Optional:                false
    DownwardAPI:             true
QoS Class:                   BestEffort
Node-Selectors:              <none>
Tolerations:                 node.kubernetes.io/not-ready:NoExecute op=Exists for 300s
                             node.kubernetes.io/unreachable:NoExecute op=Exists for 300s
Events:
  Type     Reason     Age                   From               Message
  ----     ------     ----                  ----               -------
  Normal   Scheduled  4m3s                  default-scheduler  Successfully assigned default/app-68bc7878f-4sc2q to benchmark-startup-probe-slow-app-e13361a6-control-plane
  Warning  Unhealthy  3m9s (x5 over 4m2s)   kubelet            spec.containers{app}: Liveness probe failed: Get "http://10.244.0.7:80/": dial tcp 10.244.0.7:80: connect: connection refused
  Warning  Unhealthy  3m8s (x15 over 4m3s)  kubelet            spec.containers{app}: Readiness probe failed: Get "http://10.244.0.7:80/": dial tcp 10.244.0.7:80: connect: connection refused
  Warning  BackOff    3m6s (x5 over 3m51s)  kubelet            spec.containers{app}: Back-off restarting failed container app in pod app-68bc7878f-4sc2q_default(318dabff-ce92-4cde-b503-8bf1fd7b29ff)
  Normal   Created    2m10s (x6 over 4m3s)  kubelet            spec.containers{app}: Container created
  Normal   Started    2m10s (x6 over 4m3s)  kubelet            spec.containers{app}: Container started
  Normal   Killing    2m8s (x6 over 4m2s)   kubelet            spec.containers{app}: Container app failed liveness probe, will be restarted
  Normal   Pulled     46s (x7 over 4m3s)    kubelet            spec.containers{app}: Container image "nginx:1.31.5" already present on machine and can be accessed by the pod
stderr:

Its writes:

1. `kubectl edit pod app-68bc7878f-4pfc4 --namespace default`
2. `write_file {"content": "apiVersion: v1\nkind: Pod\nmetadata:\n creationTimestamp: \"2026-10-10T01:07:19Z\"\n generateName: app-68bc7878f-\n generation: 1\n labels:\n app: app\n pod-template-hash: 68bc7878f\n name: app-68bc7878f-4pfc4\n namespace: default\n ownerReferences:\n - apiVersion: apps/v1\n blockOwnerDeletion: true\n controller: true\n kind: ReplicaSet\n name: app-68bc7878f\n uid: 979001ca-a42c-4902-9c7c-083b1073b8cb\n resourceVersion: \"903\"\n uid: 05c047c9-a6ef-45e7-837a-46ce56137224\nspec:\n containers:\n - command:\n - sh\n - -c\n - sleep 20; exec nginx -g 'daemon off;'\n image: nginx:1.31.5\n imagePullPolicy: IfNotPresent\n livenessProbe:\n failureThreshold: 1\n httpGet:\n path: /\n port: 80\n scheme: HTTP\n initialDelaySeconds: 30\n periodSeconds: 1\n successThreshold: 1\n timeoutSeconds: 1\n name: app\n ports:\n - containerPort: 80\n name: http\n protocol: TCP\n readinessProbe:\n failureThreshold: 3\n httpGet:\n path: /\n port: 80\n scheme: HTTP\n initialDelaySeconds: 30\n periodSeconds: 5\n successThreshold: 1\n timeoutSeconds: 1\n resources: {}\n terminationMessagePath: /dev/termination-log\n terminationMessagePolicy: File\n volumeMounts:\n - mountPath: /var/run/secrets/kubernetes.io/serviceaccount\n name: kube-api-access-d9hwb\n readOnly: true\n dnsPolicy: ClusterFirst\n enableServiceLinks: true\n nodeName: benchmark-startup-probe-slow-app-e13361a6-control-plane\n preemptionPolicy: PreemptLowerPriority\n priority: 0\n restartPolicy: Always\n schedulerName: default-scheduler\n securityContext: {}\n serviceAccount: default\n serviceAccountName: default\n terminationGracePeriodSeconds: 1\n tolerations:\n - effect: NoExecute\n key: node.kubernetes.io/not-ready\n operator: Exists\n tolerationSeconds: 300\n - effect: NoExecute\n key: node.kubernetes.io/unreachable\n operator: Exists\n tolerationSeconds: 300\n volumes:\n - name: kube-api-access-d9hwb\n projected:\n defaultMode: 420\n sources:\n - serviceAccountToken:\n expirationSeconds: 3607\n path: token\n - configMap:\n items:\n - key: ca.crt\n path: ca.crt\n name: kube-root-ca.crt\n - downwardAPI:\n items:\n - fieldRef:\n apiVersion: v1\n fieldPath: metadata.namespace\n path: namespace\nstatus:\n conditions:\n - lastProbeTime: null\n lastTransitionTime: \"2026-10-10T01:07:20Z\"\n observedGeneration: 1\n status: \"True\"\n type: PodReadyToStartContainers\n - lastProbeTime: null\n lastTransitionTime: \"2026-10-10T01:07:19Z\"\n observedGeneration: 1\n status: \"True\"\n type: Initialized\n - lastProbeTime: null\n lastTransitionTime: \"2026-10-10T01:07:19Z\"\n message: 'containers with unready status: [app]'\n observedGeneration: 1\n reason: ContainersNotReady\n status: \"False\"\n type: Ready\n - lastProbeTime: null\n lastTransitionTime: \"2026-10-10T01:07:19Z\"\n message: 'containers with unready status: [app]'\n observedGeneration: 1\n reason: ContainersNotReady\n status: \"False\"\n type: ContainersReady\n - lastProbeTime: null\n lastTransitionTime: \"2026-10-10T01:07:19Z\"\n observedGeneration: 1\n status: \"True\"\n type: PodScheduled\n containerStatuses:\n - containerID: containerd://d878c44f0c9e3f967bbfa9cff58f4e645d199eac800423ff36cf7d081afc88ed\n image: docker.io/library/nginx:1.31.5\n imageID: docker.io/library/nginx@sha256:05b8cb60c354a44ab824ea6e7dc69b46d50762cdbe728a347a5b656e6fb3d7c4\n lastState:\n terminated:\n containerID: containerd://d878c44f0c9e3f967bbfa9cff58f4e645d199eac800423ff36cf7d081afc88ed\n exitCode: 137\n finishedAt: \"2026-10-10T01:10:18Z\"\n reason: Error\n startedAt: \"2026-10-10T01:10:14Z\"\n name: app\n ready: false\n resources: {}\n restartCount: 6\n started: false\n state:\n waiting:\n message: back-off 2m40s restarting failed container=app pod=app-68bc7878f-4pfc4_default(05c047c9-a6ef-45e7-837a-46ce56137224)\n reason: CrashLoopBackOff\n user:\n linux:\n gid: 0\n supplementalGroups:\n - 0\n uid: 0\n volumeMounts:\n - mountPath: /var/run/secrets/kubernetes.io/serviceaccount\n name: kube-api-access-d9hwb\n readOnly: true\n hostIP: 172.19.0.7\n hostIPs:\n - ip: 172.19.0.7\n observedGeneration: 1\n phase: Running\n podIP: 10.244.0.6\n podIPs:\n - ip: 10.244.0.6\n qosClass: BestEffort\n resources: {}\n startTime: \"2026-10-10T01:07:19Z\"\n", "path": "app-pod.yaml"}`
3. `kubectl apply -f app-pod.yaml`
4. `kubectl apply -f app-pod.yaml`
5. `write_file {"content": "apiVersion: v1\nkind: Pod\nmetadata:\n creationTimestamp: \"2026-10-10T01:07:19Z\"\n generateName: app-68bc7878f-\n generation: 1\n labels:\n app: app\n pod-template-hash: 68bc7878f\n name: app-68bc7878f-4pfc4\n namespace: default\n ownerReferences:\n - apiVersion: apps/v1\n blockOwnerDeletion: true\n controller: true\n kind: ReplicaSet\n name: app-68bc7878f\n uid: 979001ca-a42c-4902-9c7c-083b1073b8cb\n resourceVersion: \"903\"\n uid: 05c047c9-a6ef-45e7-837a-46ce56137224\nspec:\n containers:\n - command:\n - sh\n - -c\n - sleep 20; exec nginx -g 'daemon off;'\n image: nginx:1.31.5\n imagePullPolicy: IfNotPresent\n livenessProbe:\n failureThreshold: 1\n httpGet:\n path: /\n port: 80\n scheme: HTTP\n initialDelaySeconds: 30\n periodSeconds: 1\n successThreshold: 1\n timeoutSeconds: 1\n name: app\n ports:\n - containerPort: 80\n name: http\n protocol: TCP\n readinessProbe:\n failureThreshold: 3\n httpGet:\n path: /\n port: 80\n scheme: HTTP\n initialDelaySeconds: 30\n periodSeconds: 5\n successThreshold: 1\n timeoutSeconds: 1\n resources: {}\n terminationMessagePath: /dev/termination-log\n terminationMessagePolicy: File\n volumeMounts:\n - mountPath: /var/run/secrets/kubernetes.io/serviceaccount\n name: kube-api-access-d9hwb\n readOnly: true\n dnsPolicy: ClusterFirst\n enableServiceLinks: true\n nodeName: benchmark-startup-probe-slow-app-e13361a6-control-plane\n preemptionPolicy: PreemptLowerPriority\n priority: 0\n restartPolicy: Always\n schedulerName: default-scheduler\n securityContext: {}\n serviceAccount: default\n serviceAccountName: default\n terminationGracePeriodSeconds: 1\n tolerations:\n - effect: NoExecute\n key: node.kubernetes.io/not-ready\n operator: Exists\n tolerationSeconds: 300\n - effect: NoExecute\n key: node.kubernetes.io/unreachable\n operator: Exists\n tolerationSeconds: 300\n volumes:\n - name: kube-api-access-d9hwb\n projected:\n defaultMode: 420\n sources:\n - serviceAccountToken:\n expirationSeconds: 3607\n path: token\n - configMap:\n items:\n - key: ca.crt\n path: ca.crt\n name: kube-root-ca.crt\n - downwardAPI:\n items:\n - fieldRef:\n apiVersion: v1\n fieldPath: metadata.namespace\n path: namespace\nstatus:\n conditions:\n - lastProbeTime: null\n lastTransitionTime: \"2026-10-10T01:07:20Z\"\n observedGeneration: 1\n status: \"True\"\n type: PodReadyToStartContainers\n - lastProbeTime: null\n lastTransitionTime: \"2026-10-10T01:07:19Z\"\n observedGeneration: 1\n status: \"True\"\n type: Initialized\n - lastProbeTime: null\n lastTransitionTime: \"2026-10-10T01:07:19Z\"\n message: 'containers with unready status: [app]'\n observedGeneration: 1\n reason: ContainersNotReady\n status: \"False\"\n type: Ready\n - lastProbeTime: null\n lastTransitionTime: \"2026-10-10T01:07:19Z\"\n message: 'containers with unready status: [app]'\n observedGeneration: 1\n reason: ContainersNotReady\n status: \"False\"\n type: ContainersReady\n - lastProbeTime: null\n lastTransitionTime: \"2026-10-10T01:07:19Z\"\n observedGeneration: 1\n status: \"True\"\n type: PodScheduled\n containerStatuses:\n - containerID: containerd://d878c44f0c9e3f967bbfa9cff58f4e645d199eac800423ff36cf7d081afc88ed\n image: docker.io/library/nginx:1.31.5\n imageID: docker.io/library/nginx@sha256:05b8cb60c354a44ab824ea6e7dc69b46d50762cdbe728a347a5b656e6fb3d7c4\n lastState:\n terminated:\n containerID: containerd://d878c44f0c9e3f967bbfa9cff58f4e645d199eac800423ff36cf7d081afc88ed\n exitCode: 137\n finishedAt: \"2026-10-10T01:10:18Z\"\n reason: Error\n startedAt: \"2026-10-10T01:10:14Z\"\n name: app\n ready: false\n resources: {}\n restartCount: 7\n started: false\n state:\n waiting:\n message: back-off 2m40s restarting failed container=app pod=app-68bc7878f-4pfc4_default(05c047c9-a6ef-45e7-837a-46ce56137224)\n reason: CrashLoopBackOff\n user:\n linux:\n gid: 0\n supplementalGroups:\n - 0\n uid: 0\n volumeMounts:\n - mountPath: /var/run/secrets/kubernetes.io/serviceaccount\n name: kube-api-access-d9hwb\n readOnly: true\n hostIP: 172.19.0.7\n hostIPs:\n - ip: 172.19.0.7\n observedGeneration: 1\n phase: Running\n podIP: 10.244.0.6\n podIPs:\n - ip: 10.244.0.6\n qosClass: BestEffort\n resources: {}\n startTime: \"2026-10-10T01:07:19Z\"\n", "path": "app-pod.yaml"}`
6. `kubectl apply -f app-pod.yaml`
7. `kubectl delete pod app-68bc7878f-4pfc4 --namespace default`
8. `kubectl apply -f app-pod.yaml`

## Teacher's write commands

1. `write_file {"content": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n annotations:\n deployment.kubernetes.io/revision: \"2\"\n creationTimestamp: \"2026-11-21T22:42:16Z\"\n generation: 2\n labels:\n app: app\n name: app\n namespace: default\n resourceVersion: \"588\"\n uid: 9e1118ae-112a-4ab1-8828-1e3d9dae881b\nspec:\n progressDeadlineSeconds: 600\n replicas: 1\n revisionHistoryLimit: 10\n selector:\n matchLabels:\n app: app\n strategy:\n type: Recreate\n template:\n metadata:\n labels:\n app: app\n spec:\n containers:\n - command:\n - sh\n - -c\n - sleep 20; exec nginx -g 'daemon off;'\n image: nginx:1.31.5\n imagePullPolicy: IfNotPresent\n livenessProbe:\n failureThreshold: 1\n httpGet:\n path: /\n port: 80\n scheme: HTTP\n initialDelaySeconds: 30\n periodSeconds: 1\n successThreshold: 1\n timeoutSeconds: 1\n name: app\n ports:\n - containerPort: 80\n name: http\n protocol: TCP\n readinessProbe:\n failureThreshold: 3\n httpGet:\n path: /\n port: 80\n scheme: HTTP\n initialDelaySeconds: 30\n periodSeconds: 5\n successThreshold: 1\n timeoutSeconds: 1\n resources: {}\n terminationMessagePath: /dev/termination-log\n terminationMessagePolicy: File\n dnsPolicy: ClusterFirst\n restartPolicy: Always\n schedulerName: default-scheduler\n securityContext: {}\n terminationGracePeriodSeconds: 1\nstatus:\n conditions:\n - lastTransitionTime: \"2026-11-21T22:42:55Z\"\n lastUpdateTime: \"2026-11-21T22:42:55Z\"\n message: Deployment does not have minimum availability.\n reason: MinimumReplicasUnavailable\n status: \"False\"\n type: Available\n - lastTransitionTime: \"2026-11-21T22:42:21Z\"\n lastUpdateTime: \"2026-11-21T22:42:57Z\"\n message: ReplicaSet \"app-68bc7878f\" is progressing.\n reason: ReplicaSetUpdated\n status: \"True\"\n type: Progressing\n observedGeneration: 2\n replicas: 1\n terminatingReplicas: 0\n unavailableReplicas: 1\n updatedReplicas: 1\n", "path": "deployment.yaml"}`
2. `kubectl apply -f deployment.yaml`

## Final report

The application Pod was failing because its startup command (`sleep 20; exec nginx ...`) takes 20 seconds to bind to port 80, but the Liveness and Readiness probes were configured with initial delay of 1s and 0s respectively. The probes were triggering immediately, failing with "connection refused," and the kubelet was killing the container before it could start, causing a `CrashLoopBackOff`.

The fix was applied to the persistent resource owning the Pod—the `Deployment/app`—by increasing the `initialDelaySeconds` of both probes to `30s`, giving the application sufficient time to initialize.

Verification of the repaired state:
*   **Deployment:** Requested replicas: 1, Updated: 1, Available: 1, Unavailable: 0. Conditions show `Available: True`.
*   **Probes:** Liveness and Readiness both now correctly show `delay=30s`.
*   **Pod:** The current Pod (`app-7598f78f7d-5gzk4`) is `1/1 Running` with `0` restarts, and its Ready condition is `True`.
