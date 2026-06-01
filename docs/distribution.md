# Distribution

## Homebrew

Use the shared `infrasecture/homebrew-tap` repository.

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

The shared tap also carries Vaka formulas. Keep myHarness updates scoped to
`Formula/myharness.rb` and `Formula/myharness-nightly.rb` so release automation
does not touch `Formula/vaka.rb` or `Formula/vaka-nightly.rb`.

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
