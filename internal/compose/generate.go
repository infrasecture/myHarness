package compose

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"github.com/infrasecture/myHarness/internal/profiles"
)

type Config struct {
	ProjectName        string
	ContainerName      string
	Workspace          string
	StateVolume        string
	Image              string
	HostUID            string
	HostGID            string
	Session            string
	Profile            profiles.Profile
	ProfileRuntimePath string
	ExtraVolumes       []string
}

func Write(w io.Writer, cfg Config) error {
	image := cfg.Image
	if image == "" {
		image = cfg.Profile.Image
	}
	session := cfg.Session
	if session == "" {
		session = cfg.Profile.DefaultSession
	}

	env := map[string]string{
		"MYHARNESS_HARNESS":  cfg.Profile.Name,
		"MYHARNESS_SESSION":  session,
		"MYHARNESS_HOST_UID": cfg.HostUID,
		"MYHARNESS_HOST_GID": cfg.HostGID,
		"MYHARNESS_HOME":     "/home/myharness",
	}
	for k, v := range cfg.Profile.Environment {
		env[k] = v
	}

	fmt.Fprintln(w, "services:")
	fmt.Fprintln(w, "  myharness:")
	fmt.Fprintf(w, "    image: %s\n", quote(image))
	fmt.Fprintf(w, "    container_name: %s\n", quote(cfg.ContainerName))
	fmt.Fprintln(w, "    restart: unless-stopped")
	fmt.Fprintln(w, "    init: true")
	fmt.Fprintln(w, "    tty: true")
	fmt.Fprintln(w, "    stdin_open: true")
	fmt.Fprintln(w, "    working_dir: /workspace")
	fmt.Fprintln(w, "    environment:")
	keys := make([]string, 0, len(env))
	for k := range env {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	for _, k := range keys {
		fmt.Fprintf(w, "      %s: %s\n", k, quote(env[k]))
	}
	fmt.Fprintln(w, "    healthcheck:")
	fmt.Fprintf(w, "      test: [\"CMD-SHELL\", %s]\n", quote(cfg.Profile.Healthcheck))
	fmt.Fprintln(w, "      interval: 1s")
	fmt.Fprintln(w, "      timeout: 3s")
	fmt.Fprintln(w, "      retries: 30")
	fmt.Fprintln(w, "      start_period: 1s")
	fmt.Fprintln(w, "    volumes:")
	fmt.Fprintf(w, "      - %s:/workspace\n", cfg.Workspace)
	fmt.Fprintln(w, "      - myharness_state:/home/myharness")
	if cfg.ProfileRuntimePath != "" {
		fmt.Fprintf(w, "      - %s\n", quote(cfg.ProfileRuntimePath+":/etc/myharness/profile.json:ro"))
	}
	for _, volume := range cfg.ExtraVolumes {
		fmt.Fprintf(w, "      - %s\n", quote(volume))
	}
	fmt.Fprintln(w, "volumes:")
	fmt.Fprintln(w, "  myharness_state:")
	fmt.Fprintf(w, "    name: %s\n", quote(cfg.StateVolume))
	fmt.Fprintln(w, "    external: true")
	return nil
}

func quote(s string) string {
	return `"` + strings.ReplaceAll(s, `"`, `\"`) + `"`
}
