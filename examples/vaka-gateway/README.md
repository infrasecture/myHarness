# Vaka Gateway Example

This example mirrors Vaka's sidecar pattern: the `myharness` workstation can
only reach the `litellm` service, and `litellm` is the only service allowed to
reach selected external model and source-control endpoints.

Run from this directory with Vaka installed:

```bash
vaka --vaka-file=vaka.yaml -f compose.yaml up -d --wait
```

The example is intentionally small. Adapt `litellm.config.yaml` and the
`services.litellm.network.egress.accept` host list before using it for real
agent traffic.
