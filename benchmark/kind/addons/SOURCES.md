# Cluster add-ons

Add-ons are installed evaluator-side by the scenarios that need them. They are
not visible to the model. Versions are pinned in the scenario `prepare` steps
and kept current by Renovate.

| Add-on           | Source (OCI)                                                    | Used by                                                                      |
| ---------------- | --------------------------------------------------------------- | ---------------------------------------------------------------------------- |
| Gateway API CRDs | `kubernetes-sigs/gateway-api` release `standard-install.yaml`   | `ingress-path-routing`, `gateway-httproute-basic`, `gateway-route-misconfig` |
| Traefik          | Helm chart OCI `ghcr.io/traefik/helm/traefik`                   | same as above                                                                |
| Calico CRDs      | Helm chart OCI `quay.io/calico/charts/crd.projectcalico.org.v1` | same as Calico                                                               |
| Calico           | Helm chart OCI `quay.io/calico/charts/tigera-operator`          | `networkpolicy-*` scenarios                                                  |

Notes:

- Helm installs reference the local pull-through proxies over plain HTTP
  (`ghcr-registry-cache:5000`, `quay-registry-cache:5000`), so chart pulls are
  cached; the version tags remain the upstream OCI tags tracked by Renovate.
- The Traefik install enables the Gateway API provider and the GatewayClass but
  disables the default Gateway, so the scenario creates its own Gateway.
- The Calico install first installs the `crd.projectcalico.org.v1` chart (the
  operator chart no longer bundles CRDs) and waits for the operator CRDs to be
  established, then installs the `tigera-operator` chart with the API server,
  Goldmane and Whisker disabled; the operator installs the dataplane from
  `quay.io`, served through the shared quay pull-through cache.
- Shared `docker.io` and `quay.io` pull-through mirrors are declared in the kind
  cluster profiles under `benchmark/kind/` via `containerdConfigPatches`; the
  in-cluster registry mirror used by the registry scenarios is in the
  specialized `cluster-registry.yaml` profile.
