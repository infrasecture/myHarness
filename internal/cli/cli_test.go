package cli

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/infrasecture/myHarness/internal/profiles"
)

func TestNormalizeVolumeRejectsManagedTargets(t *testing.T) {
	for _, spec := range []string{"./docs:/workspace", "./state:/home/myharness/"} {
		if _, err := normalizeVolume("/repo", spec); err == nil {
			t.Fatalf("expected managed target rejection for %s", spec)
		}
	}
}

func TestNormalizeVolumeResolvesRelativeSource(t *testing.T) {
	got, err := normalizeVolume("/repo/project", "./docs:/mnt/docs:ro")
	if err != nil {
		t.Fatal(err)
	}
	want := filepath.Clean("/repo/project/docs") + ":/mnt/docs:ro"
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func TestRuntimeConfigPrivateState(t *testing.T) {
	p, err := profiles.Get("codex")
	if err != nil {
		t.Fatal(err)
	}
	rt, err := newRuntime(options{workspace: "/repo/My Project", privateEnv: true}, p)
	if err != nil {
		t.Fatal(err)
	}
	if rt.project != "my-project" {
		t.Fatalf("project = %q", rt.project)
	}
	if rt.stateVolume != "my-project_myharness_state" {
		t.Fatalf("stateVolume = %q", rt.stateVolume)
	}
	if rt.container != "my-project-codex" {
		t.Fatalf("container = %q", rt.container)
	}
}

func TestRuntimeConfigHarnessState(t *testing.T) {
	p, err := profiles.Get("claude")
	if err != nil {
		t.Fatal(err)
	}
	shared, err := newRuntime(options{workspace: "/repo/app", harnessState: true}, p)
	if err != nil {
		t.Fatal(err)
	}
	if shared.stateVolume != "myharness_claude_state" {
		t.Fatalf("shared harness stateVolume = %q", shared.stateVolume)
	}

	private, err := newRuntime(options{workspace: "/repo/app", privateEnv: true, harnessState: true}, p)
	if err != nil {
		t.Fatal(err)
	}
	if private.stateVolume != "app_myharness_claude_state" {
		t.Fatalf("private harness stateVolume = %q", private.stateVolume)
	}
}

func TestResolveVakaPolicyBasicGeneratesPolicy(t *testing.T) {
	p, err := profiles.Get("codex")
	if err != nil {
		t.Fatal(err)
	}
	rt, err := newRuntime(options{workspace: "/repo/app", networkPolicy: "basic"}, p)
	if err != nil {
		t.Fatal(err)
	}
	path, ok, err := rt.resolveVakaPolicy()
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected Vaka policy to be selected")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	got := string(data)
	for _, want := range []string{"kind: ServicePolicy", "  myharness:", "- api.openai.com"} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated policy missing %q:\n%s", want, got)
		}
	}
}

func TestResolveVakaPolicyOff(t *testing.T) {
	p, err := profiles.Get("codex")
	if err != nil {
		t.Fatal(err)
	}
	rt, err := newRuntime(options{workspace: "/repo/app", networkPolicy: "off"}, p)
	if err != nil {
		t.Fatal(err)
	}
	if path, ok, err := rt.resolveVakaPolicy(); err != nil || ok || path != "" {
		t.Fatalf("resolveVakaPolicy() = %q, %v, %v; want disabled", path, ok, err)
	}
}

func TestResolveVakaPolicyPathRequiresExistingFile(t *testing.T) {
	p, err := profiles.Get("codex")
	if err != nil {
		t.Fatal(err)
	}
	rt, err := newRuntime(options{workspace: "/repo/app", networkPolicy: "path:/missing/vaka.yaml"}, p)
	if err != nil {
		t.Fatal(err)
	}
	if _, _, err := rt.resolveVakaPolicy(); err == nil {
		t.Fatal("expected missing path error")
	}
}
