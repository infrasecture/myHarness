# Distribution

## Homebrew

Use `infrasecture/tap`.

Formulas:

```text
Formula/myharness.rb
Formula/myharness-nightly.rb
```

Each formula should install:

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

The release script is designed to be extended with a tap checkout or submodule
without touching Vaka formula paths. Keep `myharness.rb` and
`myharness-nightly.rb` updates scoped to those files.

## Linux Packages

Linux packages install:

```text
/usr/local/bin/myharness
/usr/local/bin/myCodex
/usr/local/bin/myClaude
/usr/local/bin/myOpenCode
/usr/local/bin/myHermes
```

Runtime dependencies are Docker and Docker Compose v2. Packages must not bundle
Docker.

## Container Images

Publish multi-arch images for `linux/amd64` and `linux/arm64`.

Tags:

```text
<image>:latest
<image>:vX.Y.Z
<image>:<harness-version>
<image>:<harness-version>-myharness-vX.Y.Z
<image>:<git-sha>
```
