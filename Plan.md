# myHarness Refactor Plan

## Objective

Refactor the current Codex-specific container harness into `myHarness`: a
profile-driven, containerized workstation for autonomous coding agents such as
Codex, Claude, OpenCode, Hermes, and future harnesses.

Repository target:

```text
github.com/infrasecture/myHarness
```

use `gh` CLI to manage repositories.

Naming rules:

- Use `myHarness` for the project and documentation prose.
- Use `myCodex`, `myClaude`, `myOpenCode`, `myHermes`, etc. for user-facing
  compatibility aliases.
- Use `myharness` only where lowercase is required: binary names, package names,
  image names, env vars, paths, and shell commands.
- Remove Codex-specific names from generic implementation. Codex becomes one
  profile, not the organizing concept.

## Target Product

`myHarness` should provide:

- One CLI for launching and managing harness/agent workstations.
- A shared base image plus per-harness image variants.
- Optional all-in-one image for convenience.
- Persistent tmux/Byobu attach/detach sessions.
- Autonomous mode enabled inside the container.
- Non-root agent execution by default.
- Host UID/GID mapping to avoid root-owned workspace files.
- Shared or per-project persistent state.
- Extra bind mounts with managed-path conflict protection.
- Optional Vaka-powered egress control.
- GitHub workflows for CI, image publishing, packages, and releases.
- Linux and macOS distribution on par with `infrasecture/vaka`.

## Current Behavior To Preserve

The current project already has useful behavior that should become generic:

- Project-derived Compose names and container names.
- Workspace bind mount at `/workspace`.
- Persistent session created at container startup.
- `docker compose up -d --wait` readiness gating.
- `attach`, `start`, `stop`, `restart`, `ps`, `exec`, `logs`, and passthrough
  Compose management.
- Shared state and `--private-env` project state.
- Extra volume support using `source:target[:mode]`.
- Build and GHCR publishing flow.

## Architecture

### CLI Model

Implement one Go CLI:

```bash
myharness --harness codex
myharness --harness claude
myharness --harness opencode
myharness --harness hermes
```

Provide thin aliases only:

```text
myCodex    -> myharness --harness codex
myClaude   -> myharness --harness claude
myOpenCode -> myharness --harness opencode
myHermes   -> myharness --harness hermes
my*        -> myharness --harness *
```

### Image Model

Use layered images:

```text
myharness-base      common workstation tools and runtime support
myharness-codex     base + Codex CLI/config
myharness-claude    base + Claude Code/config
myharness-opencode  base + OpenCode/config
myharness-hermes    base + Hermes/config
```

Default registry names:

```text
ghcr.io/infrasecture/myharness-base
ghcr.io/infrasecture/myharness-codex
ghcr.io/infrasecture/myharness-claude
ghcr.io/infrasecture/myharness-opencode
ghcr.io/infrasecture/myharness-hermes
```

This keeps common layers shared while allowing each harness to version, test,
and publish independently.

### Repository Layout

Target layout:

```text
.
├── cmd/myharness/
├── internal/
│   ├── cli/
│   ├── compose/
│   ├── config/
│   ├── docker/
│   ├── profiles/
│   └── runtime/
├── profiles/
│   ├── codex.yaml
│   ├── claude.yaml
│   ├── opencode.yaml
│   ├── hermes.yaml
│   └── all.yaml
├── images/
│   ├── base/Dockerfile
│   ├── codex/Dockerfile
│   ├── codex/compose.yaml
│   ├── codex/profile.yaml
│   ├── claude/Dockerfile
│   ├── claude/compose.yaml
│   ├── claude/profile.yaml
│   ├── opencode/Dockerfile
│   ├── opencode/compose.yaml
│   ├── opencode/profile.yaml
│   ├── hermes/Dockerfile
│   ├── hermes/compose.yaml
│   └── hermes/profile.yaml
├── runtime/
│   ├── entrypoint.sh
│   ├── myharness-init.sh
│   └── session-banner.txt
├── bin/
│   ├── myCodex
│   ├── myClaude
│   ├── myOpenCode
│   └── myHermes
├── docs/
├── examples/
├── build.sh
├── release.sh
├── go.mod
├── README.md
```

## Profiles

Profiles define harness-specific behavior without separate launcher code.

Profile fields should include:

- Name and display name.
- Image name.
- Default command.
- Default session name.
- Required home directories.
- Environment variables.
- Config files to render into the harness home.
- Healthcheck command.
- Optional startup banner.
- Optional Vaka policy template.
- Optional upstream version resolver metadata.

Example shape:

```yaml
apiVersion: myharness.infrasecture.io/v1alpha1
kind: HarnessProfile
name: codex
displayName: myCodex
image: ghcr.io/infrasecture/myharness-codex
defaultCommand: byobu
defaultSession: codex
environment:
  CODEX_HOME: /home/myharness/.codex
config:
  files:
    - path: /home/myharness/.codex/config.toml
      mode: "0600"
      content: |
        approval_policy = "never"
        sandbox_mode = "danger-full-access"

        [projects."/workspace"]
        trust_level = "trusted"
```

Claude profile must configure Claude under `/home/myharness/.claude` and run as
non-root so autonomous mode is usable.

## Runtime And Permissions

Agent sessions must not run as root.

Startup model:

1. Container starts as root only for initialization.
2. Runtime init creates or updates the `harness` user/group.
3. UID/GID come from the invoking host user.
4. State directories are created and owned by `harness`.
5. The tmux/Byobu session and agent process run as `harness`.

Host mapping:

```text
MYHARNESS_HOST_UID = id -u
MYHARNESS_HOST_GID = id -g
container user     = harness
container home     = /home/myharness
workspace          = /workspace
```

State defaults:

```text
myharness_state
<project>_myharness_state                 with --private-env
myharness_<harness>_state                 optional harness-specific shared state
<project>_myharness_<harness>_state       optional harness-specific private state
```

Managed mount targets:

```text
/workspace
/home/myharness
```

Extra volumes must reject these targets unless a future explicit override is
added.

Ownership strategy:

- Default: run as host UID/GID.
- Chown named volumes during init when needed.
- Do not recursively chown bind-mounted workspaces by default.


## Vaka Integration

Vaka should be optional egress control, not the core permission or ownership
mechanism. When vaka.yaml is present in ~/myharness/ use vaka instead of docker compose and pass path to vaka.yaml config file.

When enabled:

- Detect `vaka`; fail with an install hint if missing.
- select a `vaka.yaml`.
- Use `vaka up` instead of `docker compose up`.


## CLI Surface

Primary commands:

```bash
myharness
myharness attach
myharness start
myharness stop
myharness restart
myharness ps
myharness logs
myharness exec bash
myharness down
myharness doctor
myharness validate
myharness version
```

Profile and image commands:

```bash
myharness profiles list
myharness profiles show codex
myharness images pull
myharness images build
myharness images build --harness codex
myharness compose config
myharness compose path
```

Common flags:

```text
--harness <name>
--profile <path>
--private-env
--volume, -v <source:target[:mode]>
--image <image>
--pull
--build
--no-build
--workspace <path>
--session <name>
--debug
```

## Compose Model

Use a generic service name:

```yaml
services:
  myharness:
    image: ghcr.io/infrasecture/myharness-codex:latest
    init: true
    tty: true
    stdin_open: true
    working_dir: /workspace
```

The CLI should generate temporary Compose overrides for:

- Container name.
- Workspace mount.
- State volume name.
- Extra volumes.
- Host UID/GID environment.
- Harness profile environment.
- Selected image.

Expose generated config for debugging:

```bash
myharness compose config
myharness compose path
```

## Build And Release

### build.sh

Create a Docker-based `build.sh` aligned with Vaka's local release workflow.

Commands:

```bash
./build.sh
./build.sh --release
./build.sh --packages
./build.sh --push
./build.sh --manifest
./build.sh --images
./build.sh --cli
./build.sh --rebuild-go
./build.sh --rebuild-images
```

Environment overrides:

```text
VERSION
ARCHS
CLI_TARGETS
REGISTRY=ghcr.io
IMAGE_PREFIX=ghcr.io/infrasecture
GOLANG_IMAGE
NFPM_IMAGE=ghcr.io/goreleaser/nfpm:latest
PUBLISH_LATEST=true
HARNESS_IMAGES="base codex claude opencode hermes"
```

Release targets:

```text
linux/amd64
linux/arm64
darwin/amd64
darwin/arm64
```

Expected outputs:

```text
dist/myharness-linux-amd64
dist/myharness-linux-arm64
dist/myharness-darwin-amd64
dist/myharness-darwin-arm64
dist/myharness_<version>_amd64.deb
dist/myharness_<version>_arm64.deb
dist/myharness-<version>-1.x86_64.rpm
dist/myharness-<version>-1.aarch64.rpm
dist/myharness-<version>-1-x86_64.pkg.tar.zst
dist/myharness-<version>-1-aarch64.pkg.tar.zst
dist/myharness-brew-darwin-amd64.tar.gz
dist/myharness-brew-darwin-arm64.tar.gz
dist/SHA256SUMS
```

Linux packages should install:

```text
/usr/local/bin/myharness
/usr/local/bin/myCodex
/usr/local/bin/myClaude
/usr/local/bin/myOpenCode
/usr/local/bin/myHermes
```

### release.sh

Create a `release.sh` matching Vaka's release style:

```bash
./release.sh
./release.sh --nightly
./release.sh --title "vX.Y.Z"
./release.sh --notes-file ./notes.md
```

Stable release behavior:

- Require a clean working tree.
- Require one `vX.Y.Z` tag on `HEAD`.
- Refuse if the GitHub release already exists.
- Run `VERSION=<tag> ./build.sh --release --packages --push`.
- Generate `SHA256SUMS`.
- Create a GitHub release with `gh`.
- Upload Linux packages, macOS binaries, Homebrew bundles, and checksums.
- Publish GHCR images and manifest lists.
- Update `infrasecture/tap`.

Nightly behavior:

- Use short commit SHA as the tag.
- Mark the release as prerelease.
- Do not update `:latest`.
- Update `myharness-nightly` Homebrew formula.

## Distribution

### Homebrew

Use `infrasecture/tap`.

Formulas:

```text
Formula/myharness.rb
Formula/myharness-nightly.rb
```

Install:

```bash
brew tap infrasecture/tap
brew install myharness
brew install myharness-nightly
```

Formula should install:

```text
myharness
myCodex
myClaude
myOpenCode
myHermes
```

Formula test:

```ruby
output = shell_output("#{bin}/myharness version")
assert_match "myharness ", output
```

Use the vaka repo aproach of tap submodule for easy interplay with release scripting. Make sure both vaka and myharness would play nicely together and the release process does not interfere with each other. 

### Linux Packages

Publish on each stable release:

- Debian/Ubuntu `.deb`.
- Fedora/RHEL/CentOS `.rpm`.
- Arch `.pkg.tar.zst`.

Examples:

```bash
curl -fLO https://github.com/infrasecture/myHarness/releases/download/v0.1.0/myharness_0.1.0_amd64.deb
sudo dpkg -i myharness_0.1.0_amd64.deb
```

```bash
curl -fLO https://github.com/infrasecture/myHarness/releases/download/v0.1.0/myharness-0.1.0-1.x86_64.rpm
sudo rpm -i myharness-0.1.0-1.x86_64.rpm
```

```bash
curl -fLO https://github.com/infrasecture/myHarness/releases/download/v0.1.0/myharness-0.1.0-1-x86_64.pkg.tar.zst
sudo pacman -U myharness-0.1.0-1-x86_64.pkg.tar.zst
```

Document Docker and Docker Compose v2 as runtime requirements. Do not package
Docker itself.

### Container Images

Publish multi-arch images for `linux/amd64` and `linux/arm64`.

Tags:

```text
<image>:latest
<image>:vX.Y.Z
<image>:<harness-version>
<image>:<harness-version>-myharness-vX.Y.Z
<image>:<git-sha> for nightly/prerelease builds
```

Examples:

```text
ghcr.io/infrasecture/myharness-codex:latest
ghcr.io/infrasecture/myharness-codex:v0.1.0
ghcr.io/infrasecture/myharness-codex:codex-0.30.1
ghcr.io/infrasecture/myharness-claude:claude-1.2.3
```

## GitHub Workflows

### CI

File:

```text
.github/workflows/ci.yml
```

Run on pull requests and pushes to `main`.

Checks:

- `gofmt`.
- `go test ./...`.
- `go vet ./...`.
- `staticcheck ./...`.
- `shellcheck build.sh release.sh bin/*`.
- YAML validation.
- Profile schema validation.
- Compose template rendering.
- Dockerfile linting.
- Native image smoke build for changed images.

### Image Build

File:

```text
.github/workflows/build-images.yml
```

Run on pull requests, pushes to `main`, and manual dispatch.

Pull requests:

- Build changed images.
- Do not push.
- Run image smoke tests.

Main/manual:

- Build selected images.
- Push only when configured by the workflow inputs or publish workflow.

### Publish Images

File:

```text
.github/workflows/publish-images.yml
```

Triggers:

```yaml
workflow_dispatch:
repository_dispatch:
  types:
    - codex_version_released
    - claude_version_released
    - opencode_version_released
    - hermes_version_released
```

Behavior:

- Resolve upstream harness versions.
- Skip tags that already exist.
- Build only required images.
- Publish multi-arch manifests.
- Update `latest` only for stable releases.

### Release

File:

```text
.github/workflows/release.yml
```

Triggers:

```yaml
workflow_dispatch:
push:
  tags:
    - "v*"
```

Behavior:

- Run release validation.
- Build CLI artifacts and packages.
- Build/publish images.
- Generate checksums.
- Create GitHub release.
- Update Homebrew tap.

Prefer calling the same `build.sh` and `release.sh` used locally so local and CI
releases stay aligned.

## Testing

### Unit Tests

Cover:

- Project name normalization.
- Alias resolution.
- Profile loading and validation.
- Compose override generation.
- Volume parsing.
- Managed mount conflict detection.
- Host UID/GID detection.
- State volume naming.
- Image tag selection.
- Vaka mode selection.
- Missing-tool error messages.

### Integration Tests

Cover:

- `myharness validate`.
- `myharness profiles list`.
- `myharness compose config`.
- Default and private state.
- Extra volumes.
- Custom image/session.
- Vaka disabled and enabled.

### Container Smoke Tests

For each image:

```bash
docker run --rm ghcr.io/infrasecture/myharness-codex:<tag> codex --version
docker run --rm ghcr.io/infrasecture/myharness-claude:<tag> claude --version
docker run --rm ghcr.io/infrasecture/myharness-opencode:<tag> opencode --version
```

Runtime checks:

- Container starts and healthcheck passes.
- Session exists.
- Agent process is non-root.
- `/workspace` is writable.
- Linux workspace writes map to host UID/GID.
- State directories are writable and correctly owned.
- Profile config files have expected ownership and mode.

### End-To-End CLI Tests

On Linux:

```bash
myharness --harness codex --workspace ./testdata/workspace start
myharness --harness codex ps
myharness --harness codex exec id
myharness --harness codex exec touch /workspace/e2e-file
myharness --harness codex stop
myharness --harness codex down
```

On macOS:

- Install through Homebrew.
- Run `myharness version`.
- Run alias smoke checks.
- Validate Docker Desktop detection.
- Start a small harness session.

When Vaka is installed:

- Validate generated `vaka.yaml`.
- Start with `--network-policy basic`.
- Confirm allowed endpoints work.
- Confirm blocked egress fails.

## Coding Style And Linting

### Go

Required:

```bash
gofmt
go test ./...
go vet ./...
staticcheck ./...
```

Guidelines:

- Keep Docker, Compose, Vaka, and Git execution behind testable interfaces.
- Generate deterministic Compose and Vaka files.
- Prefer actionable errors.
- Avoid global mutable state except build-time version metadata.

### Shell

Required:

```bash
shellcheck build.sh release.sh bin/*
```

Guidelines:

- Use `set -euo pipefail`.
- Quote variable expansions.
- Keep wrappers small.
- Fail clearly on missing tools.

### YAML And Dockerfiles

Required:

- Validate profile schema.
- Validate generated Compose YAML.
- Validate generated Vaka YAML.
- Lint Dockerfiles where practical.
- Smoke-test installed CLIs in image builds.

## Documentation

Required docs:

```text
README.md
docs/installation.md
docs/quickstart.md
docs/profiles.md
docs/security.md
docs/networking.md
docs/troubleshooting.md
docs/maintainers.md
```

Docs must explain:

- What `myHarness` is.
- Install paths for Homebrew and Linux packages.
- Harness selection and aliases.
- State behavior.
- Non-root runtime model.
- Workspace ownership expectations.
- Autonomous mode and safety model.
- Vaka integration and limits.
- Release process for maintainers.

Security docs must state:

- Mounted directories are fully available to the agent.
- The container is not a hostile-code sandbox.
- Non-root execution prevents common host permission problems but does not make
  mounted files safe.
- Vaka reduces network blast radius but does not protect mounted files.

## Migration Phases
This is not a real product migration. It describes "migration" of the current code in workdir into final solution. None of that has to be backward compatible or documented for the user. From practical point of view this is fresh rewrite.
Make sure to commit implementation as it goes, stating with the initial commit.

1. **Generic naming**
   - Introduce `myharness`.
   - Rename generic env vars to `MYHARNESS_*`.
   - Rename Compose service to `myharness`.
   - Keep old Codex env vars temporarily as deprecated aliases.
   - Keep `myCodex` as a compatibility alias.

2. **Non-root runtime**
   - Move state from `/root` to `/home/myharness`.
   - Add host UID/GID detection.
   - Create/drop to `harness` user.
   - Update Codex and Claude config paths.
   - Test Linux workspace file ownership.

3. **Profiles**
   - Add profile schema and validation.
   - Add Codex, Claude, OpenCode, Hermes profiles.
   - Generate Compose from profile data.

4. **Image split**
   - Create base and per-harness images.
   - Add smoke tests.
   - Update build scripts and workflows.

5. **Go CLI**
   - Port launcher behavior from Bash.
   - Add tests around orchestration logic.
   - Keep Bash aliases only.

6. **Distribution**
   - Add `build.sh`, `release.sh`, nfpm packages, Homebrew tap updates,
     GitHub release artifacts, and GHCR publishing.

7. **Vaka**
   - Generate Vaka policies.
   - Add sidecar gateway example and tests (akin to vaka examples).

8. **Legacy cleanup**
   - Remove Codex-specific names from generic code.
   - Remove deprecated `CODEX_*` and `MYCODEX_*` aliases after a documented
     compatibility period.

## Acceptance Criteria

The refactor is complete when:

- `myharness --harness codex` starts a Codex container with byobu and codex starts properly underneath
- `myharness --harness claude` starts the same with Claude workstation as non-root.
- `myCodex` and `myClaude` work as aliases.
- Generic Compose service is named `myharness`.
- Generic implementation no longer uses Codex as the organizing concept.
- Agent processes run as non-root host-mapped users.
- Linux workspace writes are owned by the invoking host user.
- Shared and private state modes work.
- Extra volume mounts work and reject managed target conflicts.
- GHCR publishes base, per-harness, and all-in-one images.
- Releases include Linux packages, macOS binaries, Homebrew bundles, and
  checksums.
- Homebrew install works through `infrasecture/tap`.
- CI covers tests, linting, profile validation, Compose validation, and image
  smoke tests.
- Documentation covers safety, permissions, distribution, and Vaka integration.

# Environment
You are root on host. You may install all necessary dependencies. Please document devel/build deps accordingly!

Please document leftover issues or future improvements as github issues using gh cli on the newly created infrasecture/myHarness repo.
