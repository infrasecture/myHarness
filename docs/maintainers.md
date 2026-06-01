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
`infrasecture/tap` when the tap can be cloned or a `TAP_DIR` checkout is
provided.
