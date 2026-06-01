# Networking

By default myHarness uses Docker Compose networking.

If `~/myharness/vaka.yaml` exists, startup switches to Vaka:

```bash
vaka up --config ~/myharness/vaka.yaml
```

When a Vaka policy is present and the `vaka` binary is missing, startup fails
with an install hint instead of silently falling back to unrestricted Compose
networking.

Profiles may declare a `vakaPolicyTemplate` for future policy generation. The
example policy in `examples/vaka.yaml` is intentionally minimal.
