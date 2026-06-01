package profiles

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.yaml")
	if err := os.WriteFile(path, []byte(`apiVersion: myharness.infrasecture.io/v1alpha1
kind: HarnessProfile
name: test
displayName: myTest
image: ghcr.io/infrasecture/myharness-test:latest
defaultCommand: byobu
defaultSession: test
homeDirs:
  - /home/myharness/.test
environment:
  TEST_HOME: /home/myharness/.test
config:
  files:
    - path: /home/myharness/.test/config
      mode: "0600"
      content: |
        enabled = true
healthcheck: byobu-tmux has-session -t "$${MYHARNESS_SESSION:-test}" >/dev/null 2>&1
banner: "Run test"
`), 0600); err != nil {
		t.Fatal(err)
	}

	p, err := LoadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if p.Name != "test" || p.DisplayName != "myTest" || p.Environment["TEST_HOME"] == "" {
		t.Fatalf("unexpected profile: %#v", p)
	}
	if len(p.Config.Files) != 1 || p.Config.Files[0].Path != "/home/myharness/.test/config" {
		t.Fatalf("unexpected config files: %#v", p.Config.Files)
	}
}

func TestLoadFileRejectsConfigOutsideHarnessHome(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "profile.yaml")
	if err := os.WriteFile(path, []byte(`name: bad
image: ghcr.io/infrasecture/myharness-bad:latest
defaultSession: bad
healthcheck: "true"
config:
  files:
    - path: /tmp/config
      mode: "0600"
      content: bad
`), 0600); err != nil {
		t.Fatal(err)
	}

	if _, err := LoadFile(path); err == nil {
		t.Fatal("expected config path validation error")
	}
}
