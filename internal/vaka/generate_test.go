package vaka

import (
	"strings"
	"testing"

	"github.com/infrasecture/myHarness/internal/profiles"
	"gopkg.in/yaml.v3"
)

func TestWriteGeneratesServicePolicy(t *testing.T) {
	p, err := profiles.Get("codex")
	if err != nil {
		t.Fatal(err)
	}
	var out strings.Builder
	if err := Write(&out, Config{Profile: p}); err != nil {
		t.Fatal(err)
	}
	got := out.String()
	for _, want := range []string{
		"apiVersion: agent.vaka/v1alpha1",
		"kind: ServicePolicy",
		"  myharness:",
		"block_metadata: drop",
		"- api.openai.com",
	} {
		if !strings.Contains(got, want) {
			t.Fatalf("generated policy missing %q:\n%s", want, got)
		}
	}
	var decoded map[string]any
	if err := yaml.Unmarshal([]byte(got), &decoded); err != nil {
		t.Fatalf("generated policy is not YAML: %v\n%s", err, got)
	}
}
