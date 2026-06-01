# Troubleshooting

Run diagnostics:

```bash
myharness doctor
myharness profiles validate
myharness compose config
```

Common issues:

- Docker missing: install Docker and Docker Compose v2.
- Managed mount conflict: do not mount over `/workspace` or `/home/myharness`.
- Vaka policy present but Vaka missing: install `vaka` or remove
  `~/myharness/vaka.yaml`.
- Root-owned workspace files: confirm the container receives
  `MYHARNESS_HOST_UID` and `MYHARNESS_HOST_GID` matching the host user.

The generated Compose file path is available with:

```bash
myharness compose path
```
