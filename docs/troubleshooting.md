# Troubleshooting

Run diagnostics:

```bash
myharness doctor
myharness profiles validate
myharness compose config
myharness --network-policy basic vaka config
```

Common issues:

- Docker missing: install Docker and Docker Compose v2.
- Managed mount conflict: do not mount over `/workspace` or `/home/myharness`.
- Vaka policy present but Vaka missing: install `vaka` or remove
  `~/myharness/vaka.yaml`.
- Generated Vaka policy blocks a needed endpoint: add the hostname to the
  profile's `vakaAllowedHosts` or use `--network-policy path:/path/to/vaka.yaml`
  with a project-specific policy.
- Root-owned workspace files: confirm the container receives
  `MYHARNESS_HOST_UID` and `MYHARNESS_HOST_GID` matching the host user.

The generated Compose file path is available with:

```bash
myharness compose path
myharness --network-policy basic vaka path
```
