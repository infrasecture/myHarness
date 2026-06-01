# Maintainers

## Local Checks

```bash
gofmt -w ./cmd ./internal
go test ./...
go vet ./...
shellcheck build.sh release.sh bin/*
myharness profiles validate
```

Run image builds where Docker is available:

```bash
./build.sh --images
```

Initialize the shared Homebrew tap submodule before local release work:

```bash
git submodule update --init --checkout -- homebrew-tap
```

## Releases

Stable releases require a clean working tree and exactly one `vX.Y.Z` tag on
`HEAD`.

```bash
./release.sh
```

Nightly releases:

```bash
./release.sh --nightly
```

`release.sh` creates GitHub releases through `gh`, uploads artifacts, and updates
the shared `infrasecture/homebrew-tap` through the `homebrew-tap` submodule. It
commits and pushes the tap formula first, then commits and pushes the submodule
pointer update in this repository. In GitHub Actions, set `HOMEBREW_TAP_TOKEN`
to a token with write access to the tap; locally, either use git credentials
that can push the tap or pass `TAP_TOKEN`.
