package cmd

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/pj/commandline_thing/pkg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "commandline_thing",
	Short: "Generate various things related to tmux",
	Long:  `Generate various things related to tmux`,
}

var defaultOperations = []pkg.Operation{
	&pkg.Branch{},
	&pkg.PythonVirtualEnv{},
	&pkg.VimMode{},
	&pkg.GCloudProject{},
	&pkg.ExitCode{},
}

var generatePane = &cobra.Command{
	Use:   "generate-pane",
	Short: "Generate pane content",
	Long:  `This subcommand does something specific.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return err
		}
		configFile, err := os.ReadFile(homeDir + "/.config/commandline_thing/config.json")
		if err != nil {
			return err
		}
		var config *pkg.Config
		err = json.Unmarshal(configFile, config)
		if err != nil {
			return err
		}

		nameToOperation := map[string]pkg.Operation{}
		for _, operation := range defaultOperations {
			nameToOperation[operation.Name()] = operation
		}

		operationsLeft := []pkg.Operation{}
		operationsRight := []pkg.Operation{}

		for _, operationConfig := range config.PaneConfig.OperationsLeft {
			operationName := operationConfig["name"]

			operation, ok := nameToOperation[operationName.(string)]
			if !ok {
				return fmt.Errorf("operation %s not found", operationName)
			}

			operationsLeft = append(operationsLeft, operation)
		}

		for _, operationConfig := range config.PaneConfig.OperationsRight {
			operationName := operationConfig["name"]

			operation, ok := nameToOperation[operationName.(string)]
			if !ok {
				return fmt.Errorf("operation %s not found", operationName)
			}

			operationsRight = append(operationsRight, operation)
		}

		return nil
	},
}

var runUpdates = &cobra.Command{
	Use:   "run-updates",
	Short: "trigger update of async operations",
	Run: func(cmd *cobra.Command, args []string) {

	},
}

func main() {
	rootCmd.AddCommand(generatePane)
	rootCmd.AddCommand(runUpdates)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
