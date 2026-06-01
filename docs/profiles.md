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
- `versionResolver`

Config files must render under `/home/myharness`. This prevents profiles from
overwriting managed runtime paths.
