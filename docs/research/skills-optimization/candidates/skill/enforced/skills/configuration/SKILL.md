---
name: configuration
description: Repairing ConfigMap and Secret references, applying changed configuration, and handling immutable objects.
---

# Configuration

Configuration reaches a container as environment variables or mounted files, and both are fixed when the container starts. A change at the source does not reach a running container by itself.

## Missing or wrong reference

- `CreateContainerConfigError` means the reference is wrong or a key is absent. Compare the `env`/`envFrom` and mounted-volume references with the object's actual `data` keys and name.
- Create the missing key or correct the reference (name, key, or `optional`). Do not delete the Pod in the hope it is recreated correctly.

## Applying a change

- Environment variables are read at container start. After changing a ConfigMap or Secret, trigger a new rollout and confirm the new Pods report the new value in their environment or mounted file.
- A mounted volume updates over time, but the application must reread the file; a process that cached it at start still needs a restart.

## Immutable objects

- An immutable ConfigMap or Secret rejects updates. Create a new object with the corrected content and point the workload at it, then verify; do not edit the immutable object or delete it without a replacement.

## Verify

The Pod is Ready and the running container reports the intended configuration value, read from the live Pod rather than from the source object.
