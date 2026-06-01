# Networking

By default myHarness uses Docker Compose networking.

If `~/myharness/vaka.yaml` exists, startup switches to Vaka:

```bash
vaka --vaka-file=$HOME/myharness/vaka.yaml -p <project> -f <generated-compose> up
```

When a Vaka policy is present and the `vaka` binary is missing, startup fails
with an install hint instead of silently falling back to unrestricted Compose
networking.

Network policy modes:

```bash
myharness --network-policy auto start
myharness --network-policy off start
myharness --network-policy basic start
myharness --network-policy path:/path/to/vaka.yaml start
```

`auto` is the default. It uses `~/myharness/vaka.yaml` when that file exists and
otherwise uses Docker Compose directly. `off` always uses Docker Compose.
`basic` renders the selected profile's `vakaPolicyTemplate` into a temporary
Vaka `ServicePolicy` for the generic `myharness` service. `path:<file>` uses an
explicit policy file.

Inspect the generated profile policy without starting containers:

```bash
myharness --harness codex --network-policy basic vaka config
myharness --harness codex --network-policy basic vaka path
```

Generated policies allow DNS, allow HTTPS to the selected profile's
`vakaAllowedHosts`, and drop common cloud metadata endpoints. They are intended
as a starting allowlist, not as a complete security boundary.

The `examples/vaka-gateway` directory shows a sidecar pattern where the
workstation can only reach a local LiteLLM gateway, and the gateway owns the
external model-provider allowlist.
