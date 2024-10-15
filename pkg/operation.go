package pkg

type Operation interface {
	Name() OperationName
	IsAsync() bool
	Update(string) (string, error)
	Generate(state string) (interface{}, error)
	// Load(config interface{}, state string) error
}

// Branch
type Branch struct{}

func (b *Branch) Name() OperationName                    { return "branch" }
func (b *Branch) IsAsync() bool                          { return false }
func (b *Branch) Update(string) (string, error)          { return "", nil }
func (b *Branch) Generate(_ string) (interface{}, error) { return "test", nil }

// func (b *Branch) Load(state StateStore) error {
// 	return nil
// }

// func LoadBranch(locationKey LocationKey, instanceKey InstanceKey, operationName OperationName, config interface{}) (Operation, error) {
// 	return &Branch{}, nil
// }

// // venv
// type PythonVirtualEnv struct{}

// func (*PythonVirtualEnv) Name() string                         { return "venv" }
// func (*PythonVirtualEnv) IsAsync() bool                        { return false }
// func (*PythonVirtualEnv) Show() bool                           { return true }
// func (*PythonVirtualEnv) Update(interface{}) (string, error)   { return "", nil }
// func (*PythonVirtualEnv) Generate(interface{}) (string, error) { return "", nil }

// func LoadPythonVirtualEnv(config interface{}) (Operation, error) {
// 	return &PythonVirtualEnv{}, nil
// }

// // vim mode
// type VimMode struct{}

// func (*VimMode) Name() string                         { return "vim" }
// func (*VimMode) IsAsync() bool                        { return false }
// func (*VimMode) Show() bool                           { return true }
// func (*VimMode) Update(interface{}) (string, error)   { return "", nil }
// func (*VimMode) Generate(interface{}) (string, error) { return "", nil }

// func LoadVimMode(config interface{}) (Operation, error) {
// 	return &VimMode{}, nil
// }

// // gcloud project
// type GCloudProject struct{}

// func (*GCloudProject) Name() string                         { return "gcloud" }
// func (*GCloudProject) IsAsync() bool                        { return false }
// func (*GCloudProject) Show() bool                           { return true }
// func (*GCloudProject) Update(interface{}) (string, error)   { return "", nil }
// func (*GCloudProject) Generate(interface{}) (string, error) { return "", nil }

// func LoadGCloudProject(config interface{}) (Operation, error) {
// 	return &GCloudProject{}, nil
// }

// // exit code
// type ExitCode struct{}

// func (*ExitCode) Name() string                         { return "exit_code" }
// func (*ExitCode) IsAsync() bool                        { return false }
// func (*ExitCode) Show() bool                           { return true }
// func (*ExitCode) Update(interface{}) (string, error)   { return "", nil }
// func (*ExitCode) Generate(interface{}) (string, error) { return "", nil }
// func LoadExitCode(config interface{}) (Operation, error) {
// 	return &ExitCode{}, nil
// }

type NewOperation func() Operation

type Operations map[OperationName]NewOperation

func LoadAvailableOperations() Operations {
	return map[OperationName]NewOperation{
		(&Branch{}).Name(): func() Operation { return &Branch{} },
		// (&PythonVirtualEnv{}).Name(): LoadPythonVirtualEnv,
		// (&VimMode{}).Name():          LoadVimMode,
		// (&GCloudProject{}).Name():    LoadGCloudProject,
		// (&ExitCode{}).Name():         LoadExitCode,
	}
}
