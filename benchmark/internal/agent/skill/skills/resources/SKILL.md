---
name: resources
description: Repairing CPU and memory requests and limits, OOMKilled containers and QoS class.
---

# Resources

Requests drive scheduling; limits cap runtime. Change the number that the evidence implicates, and keep every stated floor and ceiling.

## OOMKilled

- A container killed with `OOMKilled` and a restart count needs a higher memory `limit` and a matching `request` that fits the working set. Raising the limit alone, without a sane request, can still leave the Pod mis-scheduled.
- Do not remove the memory limit to stop the kill; set it to a value the workload actually fits in.

## Pending for capacity

- Compare `resources.requests` with the node's allocatable and the requests of other Pods. Reduce requests only within any stated minimum; never remove a limit to fit.

## QoS

- `Guaranteed` requires `requests == limits` for CPU and memory on **every** container, including sidecars. `Burstable` has at least one request or limit; `BestEffort` has none.
- Verify the resulting class from the live Pod: `kubectl get pod <pod> -n <ns> -o jsonpath='{.status.qosClass}'`.

## Verify

The Pod is Running and Ready with no OOM restarts, and its requests, limits and `status.qosClass` match the required outcome. Record the observed values.
