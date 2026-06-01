# Installation

## Requirements

- Docker
- Docker Compose v2
- Bash

Docker is a runtime dependency and is not packaged by myHarness.

## From Source

```bash
git clone https://github.com/infrasecture/myHarness.git
cd myHarness
./build.sh --cli
sudo install -m 0755 dist/myharness-local /usr/local/bin/myharness
sudo install -m 0755 bin/myCodex bin/myClaude bin/myOpenCode bin/myHermes /usr/local/bin/
```

## Homebrew

```bash
brew tap infrasecture/homebrew-tap
brew install myharness
```

Nightly:

```bash
brew install myharness-nightly
```

## Linux Packages

Stable releases publish `.deb`, `.rpm`, and `.pkg.tar.zst` artifacts from
GitHub releases. Packages install:

```text
/usr/local/bin/myharness
/usr/local/bin/myCodex
/usr/local/bin/myClaude
/usr/local/bin/myOpenCode
/usr/local/bin/myHermes
```
