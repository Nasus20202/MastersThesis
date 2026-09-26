# Kubernetes Images

Container images can be identified by repository, tag, or digest. A digest pins immutable image content, while a tag can be moved to different content.

`imagePullPolicy` controls image retrieval behavior. `Always` causes the kubelet to resolve the image each time a container starts, while cached layers may still be reused. `IfNotPresent` uses a locally cached image when available. `Never` requires the image to already exist on the node.

Private registries commonly use `imagePullSecrets` that reference credential Secrets in the Pod's namespace, although node-level credential providers or runtime configuration can also provide credentials.

`ErrImagePull` indicates an image retrieval attempt failed. `ImagePullBackOff` indicates Kubernetes is delaying repeated attempts after image pull failures. These states occur before the container process starts and are distinct from failures after startup.

Because image caches are node-local, a cached image can make behavior differ between nodes. Mutable tags can also produce different content over time, while digest references avoid that ambiguity.

When a tag is moved to new content but the Pod template still uses `IfNotPresent`, the kubelet can keep using the cached layers, so restarting or rolling out the workload may still run the previous build. To run the currently published content while keeping the tag, set `imagePullPolicy: Always` or reference the new digest, then confirm the running Pod reports the intended image or digest.
