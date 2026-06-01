package compose

import (
	"strings"
	"testing"

	"github.com/infrasecture/myHarness/internal/profiles"
)

func TestWriteIncludesProfileRuntimeMountAndEnvironment(t *testing.T) {
	p, err := profiles.Get("claude")
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	err = Write(&out, Config{
		ContainerName:      "project-claude",
		Workspace:          "/src/project",
		StateVolume:        "project_myharness_state",
		HostUID:            "501",
		HostGID:            "20",
		Session:            "claude",
		Profile:            p,
		ProfileRuntimePath: "/tmp/profile.json",
		ExtraVolumes:       []string{"/cache:/mnt/cache:ro"},
	})
	if err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{
		"services:\n  myharness:",
		`image: "ghcr.io/infrasecture/myharness-claude:latest"`,
		`container_name: "project-claude"`,
		`CLAUDE_CONFIG_DIR: "/home/myharness/.claude"`,
		`MYHARNESS_HOST_UID: "501"`,
		`- /src/project:/workspace`,
		`- "/tmp/profile.json:/etc/myharness/profile.json:ro"`,
		`- "/cache:/mnt/cache:ro"`,
		`name: "project_myharness_state"`,
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated compose missing %q:\n%s", want, got)
		}
	}
}
