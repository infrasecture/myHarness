package cli

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"strings"

	"github.com/infrasecture/myHarness/internal/compose"
	"github.com/infrasecture/myHarness/internal/profiles"
	"github.com/infrasecture/myHarness/internal/vaka"
)

const version = "myharness dev"

type options struct {
	harness       string
	profile       string
	privateEnv    bool
	harnessState  bool
	volumes       []string
	image         string
	pull          bool
	build         bool
	noBuild       bool
	workspace     string
	session       string
	networkPolicy string
	debug         bool
}

func Main(args []string) int {
	if alias := aliasHarness(filepath.Base(os.Args[0])); alias != "" {
		args = append([]string{"--harness", alias}, args...)
	}

	opts, rest, err := parse(args)
	if err != nil {
		fmt.Fprintln(os.Stderr, "myharness:", err)
		usage(os.Stderr)
		return 2
	}
	if opts.workspace == "" {
		opts.workspace, err = os.Getwd()
		if err != nil {
			fmt.Fprintln(os.Stderr, "myharness:", err)
			return 1
		}
	}
	opts.workspace, _ = filepath.Abs(opts.workspace)

	if len(rest) == 0 {
		rest = []string{"attach"}
	}

	cmd := rest[0]
	rest = rest[1:]
	switch cmd {
	case "help", "-h", "--help":
		usage(os.Stdout)
		return 0
	case "version":
		fmt.Println(version)
		return 0
	case "profiles":
		return profilesCmd(rest)
	}

	p, err := loadProfile(opts)
	if err != nil {
		fmt.Fprintln(os.Stderr, "myharness:", err)
		return 2
	}
	rt, err := newRuntime(opts, p)
	if err != nil {
		fmt.Fprintln(os.Stderr, "myharness:", err)
		return 2
	}

	switch cmd {
	case "attach":
		if err := rt.ensureStarted(); err != nil {
			return fail(err)
		}
		return run(rt.composeArgs("exec", "-it", "myharness", "byobu", "-r", rt.session)...)
	case "start":
		return failCode(rt.up())
	case "stop":
		return run(rt.composeArgs(append([]string{"stop"}, defaultService(rest)...)...)...)
	case "restart":
		if code := run(rt.composeArgs(append([]string{"restart"}, defaultService(rest)...)...)...); code != 0 {
			return code
		}
		return failCode(rt.up())
	case "ps":
		return run(rt.composeArgs(append([]string{"ps"}, rest...)...)...)
	case "logs":
		return run(rt.composeArgs(append([]string{"logs"}, rest...)...)...)
	case "exec":
		if len(rest) == 0 {
			fmt.Fprintln(os.Stderr, "myharness: exec requires a command")
			return 2
		}
		execArgs := append([]string{"exec"}, composeExecTTYFlags()...)
		execArgs = append(execArgs, "myharness")
		execArgs = append(execArgs, rest...)
		return run(rt.composeArgs(execArgs...)...)
	case "down":
		return run(rt.composeArgs(append([]string{"down"}, rest...)...)...)
	case "doctor":
		return doctor()
	case "validate":
		return failCode(rt.validate())
	case "compose":
		return rt.composeCmd(rest)
	case "vaka":
		return rt.vakaCmd(rest)
	case "images":
		return imagesCmd(rest, rt, opts)
	default:
		return run(rt.composeArgs(append([]string{cmd}, rest...)...)...)
	}
}

func parse(args []string) (options, []string, error) {
	var opts options
	opts.harness = "codex"
	fs := flag.NewFlagSet("myharness", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	fs.StringVar(&opts.harness, "harness", opts.harness, "")
	fs.StringVar(&opts.profile, "profile", "", "")
	fs.BoolVar(&opts.privateEnv, "private-env", false, "")
	fs.BoolVar(&opts.harnessState, "harness-state", false, "")
	fs.Func("volume", "", func(v string) error { opts.volumes = append(opts.volumes, v); return nil })
	fs.Func("v", "", func(v string) error { opts.volumes = append(opts.volumes, v); return nil })
	fs.StringVar(&opts.image, "image", "", "")
	fs.BoolVar(&opts.pull, "pull", false, "")
	fs.BoolVar(&opts.build, "build", false, "")
	fs.BoolVar(&opts.noBuild, "no-build", false, "")
	fs.StringVar(&opts.workspace, "workspace", "", "")
	fs.StringVar(&opts.session, "session", "", "")
	fs.StringVar(&opts.networkPolicy, "network-policy", "auto", "")
	fs.BoolVar(&opts.debug, "debug", false, "")
	if err := fs.Parse(args); err != nil {
		return opts, nil, err
	}
	return opts, fs.Args(), nil
}

type runtimeConfig struct {
	opts          options
	profile       profiles.Profile
	project       string
	container     string
	stateVolume   string
	composePath   string
	profilePath   string
	vakaPath      string
	workspace     string
	session       string
	normalizedVol []string
}

func newRuntime(opts options, p profiles.Profile) (*runtimeConfig, error) {
	project := projectName(opts.workspace)
	container := project + "-" + p.Name
	session := opts.session
	if session == "" {
		session = p.DefaultSession
	}
	state := "myharness_state"
	if opts.privateEnv && opts.harnessState {
		state = project + "_myharness_" + p.Name + "_state"
	} else if opts.privateEnv {
		state = project + "_myharness_state"
	} else if opts.harnessState {
		state = "myharness_" + p.Name + "_state"
	}
	if env := os.Getenv("MYHARNESS_STATE_VOLUME_NAME"); env != "" {
		state = env
	}
	vols := make([]string, 0, len(opts.volumes))
	for _, v := range opts.volumes {
		n, err := normalizeVolume(opts.workspace, v)
		if err != nil {
			return nil, err
		}
		vols = append(vols, n)
	}
	hash := sha1.Sum([]byte(opts.workspace + p.Name + strings.Join(vols, "\x00") + state + opts.image + session + opts.networkPolicy))
	path := filepath.Join(os.TempDir(), "myharness-"+hex.EncodeToString(hash[:8])+".compose.yaml")
	profilePath := filepath.Join(os.TempDir(), "myharness-"+hex.EncodeToString(hash[:8])+".profile.json")
	vakaPath := filepath.Join(os.TempDir(), "myharness-"+hex.EncodeToString(hash[:8])+".vaka.yaml")
	return &runtimeConfig{opts: opts, profile: p, project: project, container: container, stateVolume: state, composePath: path, profilePath: profilePath, vakaPath: vakaPath, workspace: opts.workspace, session: session, normalizedVol: vols}, nil
}

func (r *runtimeConfig) validate() error {
	if _, err := exec.LookPath("docker"); err != nil {
		return errors.New("docker is required")
	}
	if _, err := exec.Command("docker", "compose", "version").Output(); err != nil {
		return errors.New("docker compose v2 is required")
	}
	for _, v := range r.normalizedVol {
		parts := strings.Split(v, ":")
		if len(parts) < 2 {
			return fmt.Errorf("invalid volume %q", v)
		}
	}
	return nil
}

func (r *runtimeConfig) ensureComposeFile() error {
	if err := r.ensureProfileFile(); err != nil {
		return err
	}
	f, err := os.Create(r.composePath)
	if err != nil {
		return err
	}
	defer f.Close()
	return compose.Write(f, compose.Config{
		ProjectName:        r.project,
		ContainerName:      r.container,
		Workspace:          r.workspace,
		StateVolume:        r.stateVolume,
		Image:              r.opts.image,
		HostUID:            idOut("-u"),
		HostGID:            idOut("-g"),
		Session:            r.session,
		Profile:            r.profile,
		ProfileRuntimePath: r.profilePath,
		ExtraVolumes:       r.normalizedVol,
	})
}

func (r *runtimeConfig) ensureProfileFile() error {
	data, err := json.MarshalIndent(r.profile, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(r.profilePath, data, 0600)
}

func (r *runtimeConfig) ensureVakaFile() error {
	f, err := os.Create(r.vakaPath)
	if err != nil {
		return err
	}
	defer f.Close()
	return vaka.Write(f, vaka.Config{
		Profile:     r.profile,
		ServiceName: "myharness",
	})
}

func (r *runtimeConfig) composeArgs(args ...string) []string {
	_ = r.ensureComposeFile()
	out := []string{"docker", "compose", "-p", r.project, "-f", r.composePath}
	return append(out, args...)
}

func (r *runtimeConfig) up() error {
	if err := r.validate(); err != nil {
		return err
	}
	if err := runErr("docker", "volume", "inspect", r.stateVolume); err != nil {
		if err := runErr("docker", "volume", "create", r.stateVolume); err != nil {
			return err
		}
	}
	if r.opts.pull {
		if code := run(r.composeArgs("pull", "myharness")...); code != 0 {
			return fmt.Errorf("image pull failed")
		}
	}
	if r.opts.build && !r.opts.noBuild {
		if code := run(r.composeArgs("build", "myharness")...); code != 0 {
			return fmt.Errorf("image build failed")
		}
	}
	waitTimeout := os.Getenv("MYHARNESS_WAIT_TIMEOUT_SECONDS")
	if waitTimeout == "" {
		waitTimeout = "60"
	}
	if vakaPath, ok, err := r.resolveVakaPolicy(); err != nil {
		return err
	} else if ok {
		if _, err := exec.LookPath("vaka"); err != nil {
			return fmt.Errorf("vaka policy selected at %s but vaka is not installed", vakaPath)
		}
		args := append([]string{"--vaka-file=" + vakaPath}, r.composeGlobalArgs()...)
		args = append(args, "up", "-d", "--wait", "--wait-timeout", waitTimeout)
		return runErr("vaka", args...)
	}
	if code := run(r.composeArgs("up", "-d", "--wait", "--wait-timeout", waitTimeout)...); code != 0 {
		return fmt.Errorf("docker compose up failed")
	}
	return nil
}

func (r *runtimeConfig) resolveVakaPolicy() (string, bool, error) {
	mode := strings.TrimSpace(r.opts.networkPolicy)
	if mode == "" {
		mode = "auto"
	}
	switch mode {
	case "off", "none", "disabled":
		return "", false, nil
	case "auto":
		if path := findVakaConfig(); path != "" {
			return path, true, nil
		}
		return "", false, nil
	case "basic":
		if err := r.ensureVakaFile(); err != nil {
			return "", false, err
		}
		return r.vakaPath, true, nil
	default:
		mode = strings.TrimPrefix(mode, "path:")
		if mode == "" {
			return "", false, errors.New("--network-policy path cannot be empty")
		}
		if _, err := os.Stat(mode); err != nil {
			return "", false, fmt.Errorf("vaka policy %s: %w", mode, err)
		}
		return mode, true, nil
	}
}

func (r *runtimeConfig) ensureStarted() error {
	args := r.composeArgs("ps", "--status", "running", "--services")
	out, err := exec.Command(args[0], args[1:]...).Output()
	if err != nil || !containsLine(string(out), "myharness") {
		return r.up()
	}
	return nil
}

func (r *runtimeConfig) composeCmd(rest []string) int {
	if len(rest) == 0 {
		fmt.Fprintln(os.Stderr, "myharness: compose requires config or path")
		return 2
	}
	switch rest[0] {
	case "config":
		if err := r.ensureComposeFile(); err != nil {
			return fail(err)
		}
		data, err := os.ReadFile(r.composePath)
		if err != nil {
			return fail(err)
		}
		fmt.Print(string(data))
		return 0
	case "path":
		if err := r.ensureComposeFile(); err != nil {
			return fail(err)
		}
		fmt.Println(r.composePath)
		return 0
	default:
		return run(r.composeArgs(rest...)...)
	}
}

func (r *runtimeConfig) vakaCmd(rest []string) int {
	if len(rest) == 0 {
		fmt.Fprintln(os.Stderr, "myharness: vaka requires config or path")
		return 2
	}
	switch rest[0] {
	case "config":
		if err := r.ensureVakaFile(); err != nil {
			return fail(err)
		}
		data, err := os.ReadFile(r.vakaPath)
		if err != nil {
			return fail(err)
		}
		fmt.Print(string(data))
		return 0
	case "path":
		if err := r.ensureVakaFile(); err != nil {
			return fail(err)
		}
		fmt.Println(r.vakaPath)
		return 0
	default:
		return run(append([]string{"vaka"}, rest...)...)
	}
}

func loadProfile(opts options) (profiles.Profile, error) {
	if opts.profile != "" {
		return profiles.LoadFile(opts.profile)
	}
	return profiles.Get(opts.harness)
}

func profilesCmd(args []string) int {
	if len(args) == 0 || args[0] == "list" {
		for _, name := range profiles.Names() {
			p, _ := profiles.Get(name)
			fmt.Printf("%-10s %s %s\n", p.Name, p.DisplayName, p.Image)
		}
		return 0
	}
	if args[0] == "show" && len(args) == 2 {
		p, err := profiles.Get(args[1])
		if err != nil {
			return fail(err)
		}
		fmt.Printf("apiVersion: %s\nkind: HarnessProfile\nname: %s\ndisplayName: %s\nimage: %s\ndefaultCommand: %s\ndefaultSession: %s\n", profiles.APIVersion, p.Name, p.DisplayName, p.Image, p.DefaultCommand, p.DefaultSession)
		return 0
	}
	if args[0] == "validate" {
		paths := args[1:]
		if len(paths) == 0 {
			paths = []string{
				"profiles/codex.yaml",
				"profiles/claude.yaml",
				"profiles/opencode.yaml",
				"profiles/hermes.yaml",
				"profiles/all.yaml",
			}
		}
		for _, path := range paths {
			if _, err := profiles.LoadFile(path); err != nil {
				return fail(fmt.Errorf("%s: %w", path, err))
			}
			fmt.Printf("ok %s\n", path)
		}
		return 0
	}
	fmt.Fprintln(os.Stderr, "usage: myharness profiles list|show <name>|validate [path...]")
	return 2
}

func imagesCmd(args []string, r *runtimeConfig, opts options) int {
	if len(args) == 0 {
		fmt.Fprintln(os.Stderr, "usage: myharness images pull|build")
		return 2
	}
	switch args[0] {
	case "pull":
		return run(r.composeArgs("pull", "myharness")...)
	case "build":
		harness := opts.harness
		for i := 1; i < len(args); i++ {
			if args[i] == "--harness" && i+1 < len(args) {
				harness = args[i+1]
				i++
			}
		}
		cmd := []string{"./build.sh", "--images"}
		if harness != "" {
			cmd = append(cmd, "--harness", harness)
		}
		return run(cmd...)
	default:
		fmt.Fprintln(os.Stderr, "usage: myharness images pull|build")
		return 2
	}
}

func doctor() int {
	failures := 0
	for _, check := range [][]string{{"docker", "--version"}, {"docker", "compose", "version"}} {
		if err := runErr(check[0], check[1:]...); err != nil {
			fmt.Fprintf(os.Stderr, "missing or failing: %s\n", strings.Join(check, " "))
			failures++
		}
	}
	if failures > 0 {
		return 1
	}
	fmt.Printf("myharness doctor ok (%s/%s)\n", runtime.GOOS, runtime.GOARCH)
	return 0
}

func usage(w io.Writer) {
	fmt.Fprintln(w, `Usage:
  myharness [flags] [attach|start|stop|restart|ps|logs|exec|down|doctor|validate|version]
  myharness profiles list
  myharness profiles show codex
  myharness profiles validate
  myharness images pull
  myharness images build --harness codex
  myharness compose config
  myharness compose path
  myharness vaka config

Flags:
  --harness <name>        codex, claude, opencode, hermes, all
  --profile <path>        validate an external profile path
  --private-env           use a per-project state volume
  --harness-state         include the harness name in the state volume
  -v, --volume <spec>     add source:target[:mode] bind mount
  --image <image>         override profile image
  --pull                  pull selected image before start
  --build                 build selected compose service before start
  --no-build              disable build
  --workspace <path>      workspace to mount at /workspace
  --session <name>        tmux/byobu session name
  --network-policy <mode> off, auto, basic, or path:<vaka.yaml>
  --debug                 reserved for verbose diagnostics`)
}

func aliasHarness(name string) string {
	switch name {
	case "myCodex":
		return "codex"
	case "myClaude":
		return "claude"
	case "myOpenCode":
		return "opencode"
	case "myHermes":
		return "hermes"
	}
	return ""
}

func projectName(path string) string {
	base := strings.ToLower(filepath.Base(path))
	re := regexp.MustCompile(`[^a-z0-9_-]+`)
	name := strings.Trim(re.ReplaceAllString(base, "-"), "-")
	if name == "" {
		return "workspace"
	}
	return name
}

func normalizeVolume(workspace, spec string) (string, error) {
	if strings.ContainsAny(spec, "\n\r") {
		return "", errors.New("volume spec must not contain newlines")
	}
	parts := strings.Split(spec, ":")
	if len(parts) < 2 || len(parts) > 3 || parts[0] == "" || parts[1] == "" {
		return "", fmt.Errorf("volume spec must use source:target[:mode] syntax: %s", spec)
	}
	target := strings.TrimRight(parts[1], "/")
	if !strings.HasPrefix(target, "/") {
		return "", fmt.Errorf("volume target must be absolute: %s", spec)
	}
	if target == "/workspace" || target == "/home/myharness" {
		return "", fmt.Errorf("volume target conflicts with managed myHarness path: %s", target)
	}
	src := parts[0]
	if strings.HasPrefix(src, "~/") {
		if home, err := os.UserHomeDir(); err == nil {
			src = filepath.Join(home, strings.TrimPrefix(src, "~/"))
		}
	} else if src == "~" {
		if home, err := os.UserHomeDir(); err == nil {
			src = home
		}
	} else if strings.HasPrefix(src, ".") {
		src = filepath.Join(workspace, src)
	}
	if abs, err := filepath.Abs(src); err == nil && strings.HasPrefix(src, "/") {
		src = abs
	}
	parts[0] = src
	return strings.Join(parts, ":"), nil
}

func idOut(arg string) string {
	out, err := exec.Command("id", arg).Output()
	if err != nil {
		return "1000"
	}
	return strings.TrimSpace(string(out))
}

func defaultService(args []string) []string {
	if len(args) == 0 {
		return []string{"myharness"}
	}
	return args
}

func containsLine(text, want string) bool {
	for _, line := range strings.Split(text, "\n") {
		if strings.TrimSpace(line) == want {
			return true
		}
	}
	return false
}

func composeExecTTYFlags() []string {
	if isTerminal(os.Stdin) && isTerminal(os.Stdout) {
		return []string{"-it"}
	}
	return []string{"-T"}
}

func isTerminal(f *os.File) bool {
	info, err := f.Stat()
	if err != nil {
		return false
	}
	return info.Mode()&os.ModeCharDevice != 0
}

func findVakaConfig() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	path := filepath.Join(home, "myharness", "vaka.yaml")
	if _, err := os.Stat(path); err == nil {
		return path
	}
	return ""
}

func (r *runtimeConfig) composeGlobalArgs() []string {
	_ = r.ensureComposeFile()
	return []string{"-p", r.project, "-f", r.composePath}
}

func run(args ...string) int {
	if len(args) == 0 {
		return 0
	}
	cmd := exec.Command(args[0], args[1:]...)
	cmd.Stdin = os.Stdin
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	if err := cmd.Run(); err != nil {
		var exit *exec.ExitError
		if errors.As(err, &exit) {
			return exit.ExitCode()
		}
		fmt.Fprintln(os.Stderr, "myharness:", err)
		return 1
	}
	return 0
}

func runErr(name string, args ...string) error {
	cmd := exec.Command(name, args...)
	cmd.Stdout = os.Stdout
	cmd.Stderr = os.Stderr
	return cmd.Run()
}

func fail(err error) int {
	fmt.Fprintln(os.Stderr, "myharness:", err)
	return 1
}

func failCode(err error) int {
	if err != nil {
		return fail(err)
	}
	return 0
}
