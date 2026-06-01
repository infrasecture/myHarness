# Profiles

Profiles define harness-specific behavior without separate launcher code.

Built-ins:

```bash
myharness profiles list
myharness profiles show codex
myharness profiles validate
```

Use a custom profile:

```bash
myharness --profile ./profiles/codex.yaml compose config
```

Important fields:

- `name`
- `displayName`
- `image`
- `defaultCommand`
- `defaultSession`
- `homeDirs`
- `environment`
- `config.files`
- `healthcheck`
- `banner`
- `vakaPolicyTemplate`
- `vakaAllowedHosts`
- `versionResolver`

Config files must render under `/home/myharness`. This prevents profiles from
overwriting managed runtime paths.

`vakaPolicyTemplate` is a Go text/template that renders a Vaka `ServicePolicy`.
The built-in template receives `ServiceName`, `HarnessName`, and
`AllowedHosts`. `vakaAllowedHosts` should contain the HTTPS hostnames needed by
that harness profile.
