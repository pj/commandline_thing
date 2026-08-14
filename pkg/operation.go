package pkg

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type Operation interface {
	Name() OperationName
	IsAsync() bool
	Update(string, string) (string, error)
	Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error)
}

// Configurable is implemented by operations that take extra fields from
// their YAML entry beyond `type` (e.g. cycle's `name`/`names`). rawConfig is
// the entry's full raw map, as decoded from YAML — including `type` itself.
type Configurable interface {
	Configure(rawConfig map[string]interface{}) error
}

// Git
type Git struct{}

type GitResult struct {
	Branch string
	Status string
}

func (b *Git) Name() OperationName                   { return "git" }
func (b *Git) IsAsync() bool                         { return false }
func (b *Git) Update(_ string, state string) (string, error) { return state, nil }
func (b *Git) Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error) {
	cmd := exec.Command("git", "rev-parse", "--abbrev-ref", "HEAD")
	cmd.Dir = locationPath
	output, err := cmd.Output()
	if err != nil {
		return nil, nil
	}

	branch := strings.TrimSpace(string(output))

	commandStatus := exec.Command("git", "status", "-s")
	commandStatus.Dir = locationPath
	outputStatus, err := commandStatus.Output()
	if err != nil {
		return nil, nil
	}

	status := strings.TrimSpace(string(outputStatus))

	return GitResult{Branch: branch, Status: status}, nil
}

// venv
type PythonVirtualEnv struct{}

func (*PythonVirtualEnv) Name() OperationName                   { return "venv" }
func (*PythonVirtualEnv) IsAsync() bool                         { return false }
func (*PythonVirtualEnv) Update(_ string, state string) (string, error) { return state, nil }
func (*PythonVirtualEnv) Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error) {
	return state, nil
}

// vim mode
type VimMode struct{}

func (*VimMode) Name() OperationName                   { return "vim" }
func (*VimMode) IsAsync() bool                         { return false }
func (*VimMode) Update(_ string, state string) (string, error) { return state, nil }
func (*VimMode) Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error) {
	return state, nil
}

// // gcloud project
type GCloudProject struct{}

func (*GCloudProject) Name() OperationName                   { return "gcloud" }
func (*GCloudProject) IsAsync() bool                         { return false }
func (*GCloudProject) Update(_ string, state string) (string, error) { return state, nil }
func (*GCloudProject) Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error) {
	// #  if type "gcloud" > /dev/null && gcloud projects list > /dev/null 2>&1 ; then
	// #    tmux setenv -g "PANE_GCLOUD_PROJECT${IDS}" "$(gcloud config get-value project)"
	// #    tmux refresh-client -S
	// #  fi

	return "", nil
}

// // exit code
type ExitCode struct{}

func (*ExitCode) Name() OperationName                   { return "exit_code" }
func (*ExitCode) IsAsync() bool                         { return false }
func (*ExitCode) Update(_ string, state string) (string, error) { return state, nil }
func (*ExitCode) Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error) {
	return state, nil
}

type WorkingDirectory struct{}

func (*WorkingDirectory) Name() OperationName                   { return "working_directory" }
func (*WorkingDirectory) IsAsync() bool                         { return false }
func (*WorkingDirectory) Update(_ string, state string) (string, error) { return state, nil }
func (*WorkingDirectory) Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return "", err
	}

	if strings.HasPrefix(locationPath, homeDir) {
		correctedPath := strings.Replace(locationPath, homeDir, "~", 1)
		return correctedPath, nil
	}

	return locationPath, nil
}

type TmuxActivePane struct{}

func (*TmuxActivePane) Name() OperationName                   { return "tmux_active_pane" }
func (*TmuxActivePane) IsAsync() bool                         { return false }
func (*TmuxActivePane) Update(_ string, state string) (string, error) { return state, nil }
func (*TmuxActivePane) Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error) {
	tmux := os.Getenv("TMUX")
	if tmux == "" {
		return false, nil
	}

	paneId := strings.Split(string(instanceKey), ".")[1]

	cmd := exec.Command("tmux", "display", "-p", "#{=-1:pane_id}")
	cmd.Dir = locationPath
	output, err := cmd.Output()
	if err != nil {
		return false, err
	}

	return strings.TrimSpace(string(output)) == paneId, nil
}

type TmuxCurrentPane struct{}

func (*TmuxCurrentPane) Name() OperationName                   { return "tmux_current_pane" }
func (*TmuxCurrentPane) IsAsync() bool                         { return false }
func (*TmuxCurrentPane) Update(_ string, state string) (string, error) { return state, nil }
func (*TmuxCurrentPane) Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error) {
	tmux := os.Getenv("TMUX")
	if tmux == "" {
		return "", nil
	}
	paneId := strings.Split(string(instanceKey), ".")[1]
	return paneId, nil
}

type InTmux struct{}

func (*InTmux) Name() OperationName                   { return "in_tmux" }
func (*InTmux) IsAsync() bool                         { return false }
func (*InTmux) Update(_ string, state string) (string, error) { return state, nil }
func (*InTmux) Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error) {
	tmux := os.Getenv("TMUX")
	return tmux != "", nil
}

type HostDetails struct{}

type HostDetailsResult struct {
	Hostname string
	IsSSH    bool
}

func (*HostDetails) Name() OperationName                   { return "host_details" }
func (*HostDetails) IsAsync() bool                         { return false }
func (*HostDetails) Update(_ string, state string) (string, error) { return state, nil }
func (*HostDetails) Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return HostDetailsResult{Hostname: hostname, IsSSH: false}, nil
}

// Meme exposes every meme in the meme directory to templates as
// .meme.<name>, each as a single character at that meme's PUA code point
// (see MemeCodepoint) — normal cell content that survives tmux redraws,
// unlike an OSC 1337 inline image (which tmux doesn't track in its own
// screen buffer, so it vanishes on the next redraw). Rendering requires
// MemeTerminal with MemeFont.ttf installed; other terminals show tofu.
// Names containing characters that aren't valid in a template's dotted
// field access (e.g. "doge-2") need {{ index .meme "doge-2" }} instead of
// {{ .meme.doge-2 }}.
type Meme struct{}

func (*Meme) Name() OperationName                   { return "meme" }
func (*Meme) IsAsync() bool                         { return false }
func (*Meme) Update(_ string, state string) (string, error) { return state, nil }
func (*Meme) Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error) {
	dir := MemeDir()
	names, err := ListMemes(dir)
	if err != nil {
		return nil, err
	}

	memes := make(map[string]string, len(names))
	for i, name := range names {
		memes[name] = string(rune(MemeCodepointBase + i))
	}

	return memes, nil
}

// Cycle steps through a configured, ordered list of meme names by one on
// every Update() call (wrapping around), and Generate() returns whichever
// name is current. It's a name lookup, not a rendered character — combine
// it with Meme's output in a template via {{ index .meme .<name> }} to get
// the actual glyph.
//
// Configured in YAML as:
//
//	- type: cycle
//	  name: nyan
//	  names: [nyan1, nyan2, nyan3, nyan4]
//
// `name` becomes this instance's effective Name() — the template field
// (.nyan) and the state-store key both key off it instead of the fixed
// "cycle" type name, so multiple cycle operations can coexist in one
// location's operations list without colliding, each with independent
// state. Something still needs to actually call Update() for state to
// advance each render — Generate() alone only reads the current index; see
// `commandline_thing update`.
type Cycle struct {
	name  string
	names []string
}

func (c *Cycle) Name() OperationName {
	if c.name != "" {
		return OperationName(c.name)
	}
	return "cycle"
}
func (*Cycle) IsAsync() bool { return false }

func (c *Cycle) Configure(rawConfig map[string]interface{}) error {
	if nameRaw, ok := rawConfig["name"]; ok {
		name, ok := nameRaw.(string)
		if !ok {
			return fmt.Errorf("cycle: name must be a string")
		}
		c.name = name
	}

	namesRaw, ok := rawConfig["names"]
	if !ok {
		return fmt.Errorf("cycle: names is required")
	}
	namesList, ok := namesRaw.([]interface{})
	if !ok {
		return fmt.Errorf("cycle: names must be a list")
	}
	names := make([]string, 0, len(namesList))
	for _, n := range namesList {
		s, ok := n.(string)
		if !ok {
			return fmt.Errorf("cycle: names must be a list of strings")
		}
		names = append(names, s)
	}
	if len(names) == 0 {
		return fmt.Errorf("cycle: names must not be empty")
	}
	c.names = names
	return nil
}

func (c *Cycle) currentIndex(state string) int {
	idx, err := strconv.Atoi(state)
	if err != nil || idx < 0 || idx >= len(c.names) {
		return 0
	}
	return idx
}

func (c *Cycle) Update(locationPath string, state string) (string, error) {
	next := (c.currentIndex(state) + 1) % len(c.names)
	return strconv.Itoa(next), nil
}

func (c *Cycle) Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error) {
	return c.names[c.currentIndex(state)], nil
}

type NewOperation func() Operation

type Operations map[OperationName]NewOperation

func LoadAvailableOperations() Operations {
	return map[OperationName]NewOperation{
		(&Git{}).Name():              func() Operation { return &Git{} },
		(&PythonVirtualEnv{}).Name(): func() Operation { return &PythonVirtualEnv{} },
		(&VimMode{}).Name():          func() Operation { return &VimMode{} },
		(&GCloudProject{}).Name():    func() Operation { return &GCloudProject{} },
		(&ExitCode{}).Name():         func() Operation { return &ExitCode{} },
		(&WorkingDirectory{}).Name(): func() Operation { return &WorkingDirectory{} },
		(&TmuxActivePane{}).Name():   func() Operation { return &TmuxActivePane{} },
		(&TmuxCurrentPane{}).Name():  func() Operation { return &TmuxCurrentPane{} },
		(&HostDetails{}).Name():      func() Operation { return &HostDetails{} },
		(&InTmux{}).Name():           func() Operation { return &InTmux{} },
		(&Meme{}).Name():             func() Operation { return &Meme{} },
		(&Cycle{}).Name():            func() Operation { return &Cycle{} },
	}
}
