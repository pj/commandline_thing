package pkg

type Operation interface {
	Name() string
	IsAsync() bool
	Show() bool
	Run(interface{}) (string, error)
	Output(interface{}) (string, error)
}

// Branch

type Branch struct{}

func (*Branch) Name() string                       { return "branch" }
func (*Branch) IsAsync() bool                      { return false }
func (*Branch) Show() bool                         { return true }
func (*Branch) Run(interface{}) (string, error)    { return "", nil }
func (*Branch) Output(interface{}) (string, error) { return "", nil }

// venv
type PythonVirtualEnv struct{}

func (*PythonVirtualEnv) Name() string                       { return "venv" }
func (*PythonVirtualEnv) IsAsync() bool                      { return false }
func (*PythonVirtualEnv) Show() bool                         { return true }
func (*PythonVirtualEnv) Run(interface{}) (string, error)    { return "", nil }
func (*PythonVirtualEnv) Output(interface{}) (string, error) { return "", nil }

// vim mode
type VimMode struct{}

func (*VimMode) Name() string                       { return "vim" }
func (*VimMode) IsAsync() bool                      { return false }
func (*VimMode) Show() bool                         { return true }
func (*VimMode) Run(interface{}) (string, error)    { return "", nil }
func (*VimMode) Output(interface{}) (string, error) { return "", nil }

// gcloud project
type GCloudProject struct{}

func (*GCloudProject) Name() string                       { return "gcloud" }
func (*GCloudProject) IsAsync() bool                      { return false }
func (*GCloudProject) Show() bool                         { return true }
func (*GCloudProject) Run(interface{}) (string, error)    { return "", nil }
func (*GCloudProject) Output(interface{}) (string, error) { return "", nil }

// exit code
type ExitCode struct{}

func (*ExitCode) Name() string                       { return "exit_code" }
func (*ExitCode) IsAsync() bool                      { return false }
func (*ExitCode) Show() bool                         { return true }
func (*ExitCode) Run(interface{}) (string, error)    { return "", nil }
func (*ExitCode) Output(interface{}) (string, error) { return "", nil }
