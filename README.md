# myHarness

myHarness is a profile-driven, containerized workstation for autonomous coding
agents such as Codex, Claude Code, OpenCode, Hermes, and future harnesses.

It provides one CLI, persistent Byobu/tmux sessions, non-root agent execution,
host UID/GID mapping, per-project or shared state, managed extra bind mounts,
and layered container images.

Repository target: `github.com/infrasecture/myHarness`

## Requirements

- Docker
- Docker Compose v2
- Bash
- Go 1.22 or a built `myharness` binary

Vaka egress control is optional. If `~/myharness/vaka.yaml` exists, `myharness`
uses `vaka --vaka-file=$HOME/myharness/vaka.yaml up` for startup and requires the
`vaka` CLI to be installed. You can also request a generated profile policy
with `--network-policy basic`.

## Quick Start

Build the local CLI:

```bash
./build.sh --cli
```

Start the default Codex workstation for the current directory:

```bash
bin/myharness
```

Select another harness:

```bash
bin/myharness --harness claude
bin/myharness --harness opencode
bin/myharness --harness hermes
```

Compatibility aliases are thin wrappers:

```bash
bin/myCodex
bin/myClaude
bin/myOpenCode
bin/myHermes
```

## CLI

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
myharness images build --harness codex
myharness compose config
myharness compose path
myharness vaka config
myharness vaka path
```

Common flags:

```text
--harness <name>
--profile <path>
--private-env
--harness-state
--volume, -v <source:target[:mode]>
--image <image>
--pull
--build
--no-build
--workspace <path>
--session <name>
--network-policy <off|auto|basic|path:vaka.yaml>
--debug
```

## Runtime Model

Containers start as root only for initialization. The entrypoint creates or
updates a `harness` user and group using:

```text
MYHARNESS_HOST_UID = id -u
MYHARNESS_HOST_GID = id -g
container user     = harness
container home     = /home/myharness
workspace          = /workspace
```

The Byobu/tmux session and agent process run as `harness`. The workspace bind
mount is not recursively chowned.

State defaults:

```text
myharness_state
<project>_myharness_state with --private-env
myharness_<harness>_state with --harness-state
<project>_myharness_<harness>_state with --private-env --harness-state
```

Managed mount targets are protected:

```text
/workspace
/home/myharness
```

Extra volumes targeting those paths are rejected.

## Images

Layered images:

```text
ghcr.io/infrasecture/myharness-base
ghcr.io/infrasecture/myharness-codex
ghcr.io/infrasecture/myharness-claude
ghcr.io/infrasecture/myharness-opencode
ghcr.io/infrasecture/myharness-hermes
```

Build images locally:

```bash
./build.sh --images
./build.sh --images --harness codex
```

## Build And Release

Build CLI artifacts:

```bash
./build.sh --cli
```

Build release artifacts:

```bash
VERSION=v0.1.0 ./build.sh --release --packages --images
```

Release with GitHub CLI:

```bash
./release.sh
./release.sh --nightly
./release.sh --title "v0.1.0" --notes-file ./notes.md
```

Stable releases require a clean working tree and exactly one `vX.Y.Z` tag on
`HEAD`. Nightly releases use the short commit SHA and do not update `:latest`.

## Repository Layout

```text
cmd/myharness/          Go CLI entrypoint
internal/               CLI, Compose, profile, and runtime packages
profiles/               Built-in profile documents
images/                 Base and per-harness Dockerfiles
runtime/                Container entrypoint and session initialization
bin/                    CLI wrappers and compatibility aliases
docs/                   Distribution and operational docs
examples/               Vaka policy template and gateway example
build.sh                Local build workflow
release.sh              GitHub release workflow
```

## More Docs

- [Installation](docs/installation.md)
- [Quickstart](docs/quickstart.md)
- [Profiles](docs/profiles.md)
- [Security](docs/security.md)
- [Networking](docs/networking.md)
- [Troubleshooting](docs/troubleshooting.md)
- [Maintainers](docs/maintainers.md)
