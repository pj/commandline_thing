package pkg

import (
	"os"
	"os/exec"
	"strings"
)

type Operation interface {
	Name() OperationName
	IsAsync() bool
	Update(string, string) (string, error)
	Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error)
}

// Git
type Git struct{}

type GitResult struct {
	Branch string
	Status string
}

func (b *Git) Name() OperationName                   { return "git" }
func (b *Git) IsAsync() bool                         { return false }
func (b *Git) Update(string, string) (string, error) { return "", nil }
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
func (*PythonVirtualEnv) Update(string, string) (string, error) { return "", nil }
func (*PythonVirtualEnv) Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error) {
	return state, nil
}

// vim mode
type VimMode struct{}

func (*VimMode) Name() OperationName                   { return "vim" }
func (*VimMode) IsAsync() bool                         { return false }
func (*VimMode) Update(string, string) (string, error) { return "", nil }
func (*VimMode) Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error) {
	return state, nil
}

// // gcloud project
type GCloudProject struct{}

func (*GCloudProject) Name() OperationName                   { return "gcloud" }
func (*GCloudProject) IsAsync() bool                         { return false }
func (*GCloudProject) Update(string, string) (string, error) { return "", nil }
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
func (*ExitCode) Update(string, string) (string, error) { return "", nil }
func (*ExitCode) Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error) {
	return state, nil
}

type WorkingDirectory struct{}

func (*WorkingDirectory) Name() OperationName                   { return "working_directory" }
func (*WorkingDirectory) IsAsync() bool                         { return false }
func (*WorkingDirectory) Update(string, string) (string, error) { return "", nil }
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
func (*TmuxActivePane) Update(string, string) (string, error) { return "", nil }
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
func (*TmuxCurrentPane) Update(string, string) (string, error) { return "", nil }
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
func (*InTmux) Update(string, string) (string, error) { return "", nil }
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
func (*HostDetails) Update(string, string) (string, error) { return "", nil }
func (*HostDetails) Generate(locationKey LocationKey, instanceKey InstanceKey, locationPath string, state string) (interface{}, error) {
	hostname, err := os.Hostname()
	if err != nil {
		return nil, err
	}

	return HostDetailsResult{Hostname: hostname, IsSSH: false}, nil
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
	}
}
