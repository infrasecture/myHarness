package profiles

import (
	"fmt"
	"os"
	"sort"
	"strings"

	"gopkg.in/yaml.v3"
)

type File struct {
	Path    string `json:"path" yaml:"path"`
	Mode    string `json:"mode" yaml:"mode"`
	Content string `json:"content" yaml:"content"`
}

type Profile struct {
	APIVersion              string            `json:"apiVersion" yaml:"apiVersion"`
	Kind                    string            `json:"kind" yaml:"kind"`
	Name                    string            `json:"name" yaml:"name"`
	DisplayName             string            `json:"displayName" yaml:"displayName"`
	Image                   string            `json:"image" yaml:"image"`
	DefaultCommand          string            `json:"defaultCommand" yaml:"defaultCommand"`
	DefaultSession          string            `json:"defaultSession" yaml:"defaultSession"`
	HomeDirs                []string          `json:"homeDirs" yaml:"homeDirs"`
	Environment             map[string]string `json:"environment" yaml:"environment"`
	Config                  Config            `json:"config" yaml:"config"`
	Healthcheck             string            `json:"healthcheck" yaml:"healthcheck"`
	Banner                  string            `json:"banner" yaml:"banner"`
	VakaPolicyTemplate      string            `json:"vakaPolicyTemplate,omitempty" yaml:"vakaPolicyTemplate"`
	UpstreamVersionResolver map[string]string `json:"versionResolver,omitempty" yaml:"versionResolver"`
}

type Config struct {
	Files []File `json:"files" yaml:"files"`
}

const APIVersion = "myharness.infrasecture.io/v1alpha1"

var builtins = map[string]Profile{
	"codex": {
		Name:           "codex",
		DisplayName:    "myCodex",
		Image:          "ghcr.io/infrasecture/myharness-codex:latest",
		DefaultCommand: "byobu",
		DefaultSession: "codex",
		HomeDirs:       []string{"/home/myharness/.codex"},
		Environment: map[string]string{
			"CODEX_HOME": "/home/myharness/.codex",
		},
		Config: Config{
			Files: []File{{
				Path: "/home/myharness/.codex/config.toml",
				Mode: "0600",
				Content: `approval_policy = "never"
sandbox_mode = "danger-full-access"

[projects."/workspace"]
trust_level = "trusted"
`,
			}},
		},
		Healthcheck: "byobu-tmux has-session -t \"$${MYHARNESS_SESSION:-codex}\" >/dev/null 2>&1",
		Banner:      "Run Codex with: codex",
	},
	"claude": {
		Name:           "claude",
		DisplayName:    "myClaude",
		Image:          "ghcr.io/infrasecture/myharness-claude:latest",
		DefaultCommand: "byobu",
		DefaultSession: "claude",
		HomeDirs:       []string{"/home/myharness/.claude"},
		Environment: map[string]string{
			"CLAUDE_CONFIG_DIR": "/home/myharness/.claude",
		},
		Config: Config{
			Files: []File{{
				Path: "/home/myharness/.claude/settings.json",
				Mode: "0600",
				Content: `{
  "$schema": "https://json.schemastore.org/claude-code-settings.json",
  "permissions": {
    "defaultMode": "bypassPermissions",
    "skipDangerousModePermissionPrompt": true
  }
}
`,
			}},
		},
		Healthcheck: "byobu-tmux has-session -t \"$${MYHARNESS_SESSION:-claude}\" >/dev/null 2>&1",
		Banner:      "Run Claude Code with: claude",
	},
	"opencode": {
		Name:           "opencode",
		DisplayName:    "myOpenCode",
		Image:          "ghcr.io/infrasecture/myharness-opencode:latest",
		DefaultCommand: "byobu",
		DefaultSession: "opencode",
		HomeDirs:       []string{"/home/myharness/.config/opencode"},
		Environment:    map[string]string{},
		Healthcheck:    "byobu-tmux has-session -t \"$${MYHARNESS_SESSION:-opencode}\" >/dev/null 2>&1",
		Banner:         "Run OpenCode with: opencode",
	},
	"hermes": {
		Name:           "hermes",
		DisplayName:    "myHermes",
		Image:          "ghcr.io/infrasecture/myharness-hermes:latest",
		DefaultCommand: "byobu",
		DefaultSession: "hermes",
		HomeDirs:       []string{"/home/myharness/.hermes"},
		Environment:    map[string]string{},
		Healthcheck:    "byobu-tmux has-session -t \"$${MYHARNESS_SESSION:-hermes}\" >/dev/null 2>&1",
		Banner:         "Run Hermes from this session.",
	},
	"all": {
		Name:           "all",
		DisplayName:    "myHarness All",
		Image:          "ghcr.io/infrasecture/myharness-all:latest",
		DefaultCommand: "byobu",
		DefaultSession: "myharness",
		HomeDirs:       []string{"/home/myharness/.codex", "/home/myharness/.claude", "/home/myharness/.config/opencode", "/home/myharness/.hermes"},
		Environment: map[string]string{
			"CODEX_HOME":        "/home/myharness/.codex",
			"CLAUDE_CONFIG_DIR": "/home/myharness/.claude",
		},
		Healthcheck: "byobu-tmux has-session -t \"$${MYHARNESS_SESSION:-myharness}\" >/dev/null 2>&1",
		Banner:      "Codex, Claude Code, OpenCode, and Hermes tooling are available.",
	},
}

func Get(name string) (Profile, error) {
	name = strings.ToLower(strings.TrimSpace(name))
	p, ok := builtins[name]
	if !ok {
		return Profile{}, fmt.Errorf("unknown harness %q", name)
	}
	applyDefaults(&p)
	return p, nil
}

func LoadFile(path string) (Profile, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return Profile{}, err
	}
	var p Profile
	if err := yaml.Unmarshal(data, &p); err != nil {
		return Profile{}, err
	}
	if err := validate(&p); err != nil {
		return Profile{}, err
	}
	applyDefaults(&p)
	return p, nil
}

func validate(p *Profile) error {
	if p.Name == "" {
		return fmt.Errorf("profile missing name")
	}
	if p.Image == "" {
		return fmt.Errorf("profile %q missing image", p.Name)
	}
	if p.DefaultSession == "" {
		return fmt.Errorf("profile %q missing defaultSession", p.Name)
	}
	if p.Healthcheck == "" {
		return fmt.Errorf("profile %q missing healthcheck", p.Name)
	}
	for _, f := range p.Config.Files {
		if !strings.HasPrefix(f.Path, "/home/myharness/") {
			return fmt.Errorf("profile %q config path must be under /home/myharness: %s", p.Name, f.Path)
		}
		if f.Mode == "" {
			return fmt.Errorf("profile %q config file %s missing mode", p.Name, f.Path)
		}
	}
	return nil
}

func applyDefaults(p *Profile) {
	if p.APIVersion == "" {
		p.APIVersion = APIVersion
	}
	if p.Kind == "" {
		p.Kind = "HarnessProfile"
	}
	if p.DisplayName == "" {
		p.DisplayName = p.Name
	}
	if p.DefaultCommand == "" {
		p.DefaultCommand = "byobu"
	}
	if p.Environment == nil {
		p.Environment = map[string]string{}
	}
}

func Names() []string {
	names := make([]string, 0, len(builtins))
	for name := range builtins {
		names = append(names, name)
	}
	sort.Strings(names)
	return names
}
