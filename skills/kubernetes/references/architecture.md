# Kubernetes architecture and control loops

Kubernetes is a set of asynchronous components. The API server authenticates,
authorizes, validates and persists API objects; admission may mutate or reject
them. Controllers reconcile resources, the scheduler assigns unscheduled Pods,
and the kubelet asks the container runtime to execute Pods on nodes. Add-ons can
provide networking, storage, metrics or custom controllers.

Localize a problem before changing anything:

- An operation rejected by the API or admission layer is different from an
  object that was accepted but has not been reconciled.
- A controller's missing or stale child object points to ownership, controller
  health or dependency evidence; the scheduler is not responsible for creating
  workloads.
- A scheduled Pod that cannot start belongs to the node, runtime, image,
  volume, security or container-lifecycle path rather than the scheduler.
- A running but unready Pod requires application, probe, configuration or
  network evidence; node readiness alone does not establish service health.

Correlate object generations, UIDs, status conditions, events and timestamps.
Do not infer a control-plane or add-on failure from one stale observation, and
do not inspect or modify underlying control-plane data stores unless the task
explicitly provides that operation.
