package main

import (
	"fmt"
	"log"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pj/commandline_thing/pkg"
	"github.com/spf13/cobra"
)

func setup(locationKey pkg.LocationKey) (*pkg.Location, pkg.StateStore, *pkg.AllConfigs, error) {
	availableOperations := pkg.LoadAvailableOperations()

	config, err := pkg.LoadConfig(availableOperations)

	if err != nil {
		return nil, nil, nil, err
	}

	locationConfig, ok := config.Configs[locationKey]
	if !ok {
		return nil, nil, nil, fmt.Errorf("config not found: %s", locationKey)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to get user's home directory: %w", err)
	}

	stateDBPath := filepath.Join(homeDir, ".config", "commandline_thing", "state.db")

	state, err := pkg.NewSQLiteState(stateDBPath)
	if err != nil {
		return nil, nil, nil, fmt.Errorf("failed to create state: %w", err)
	}

	return &locationConfig, state, config, nil
}

func setupLogger() (*log.Logger, error) {
	// Set up logging to file
	logDir := filepath.Join(os.Getenv("HOME"), ".config", "commandline_thing")
	if err := os.MkdirAll(logDir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create log directory: %w", err)
	}

	logFile, err := os.OpenFile(filepath.Join(logDir, "command.log"), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, fmt.Errorf("failed to open log file: %w", err)
	}

	logger := log.New(logFile, "", log.LstdFlags|log.Lshortfile)
	return logger, nil
}

func runPostCommands(config *pkg.AllConfigs, logger *log.Logger) error {
	for _, postCommand := range config.PostCommands {
		cmd := exec.Command("sh", "-c", postCommand)
		logger.Printf("running post command: %s", postCommand)
		cmd.Stdout = logger.Writer()
		cmd.Stderr = logger.Writer()
		err := cmd.Start()
		if err != nil {
			logger.Printf("failed to run post command %s: %s", postCommand, err)
			return err
		}
	}

	return nil
}

func main() {
	var rootCmd = &cobra.Command{
		Use:   "commandline_thing",
		Short: "Generate various things related to tmux",
		Long:  `Generate various things related to tmux`,
	}

	var generateCmd = &cobra.Command{
		Use:   "generate",
		Short: `Generate content for a location`,
		RunE: func(cmd *cobra.Command, args []string) error {
			logger, err := setupLogger()
			if err != nil {
				fmt.Println("failed to setup logger: %s", err)
				return err
			}

			locationKey := pkg.LocationKey(args[0])
			locationConfig, stateStore, _, err := setup(locationKey)
			if err != nil {
				logger.Printf("failed to setup: %s", err)
				return err
			}

			instanceKey := pkg.InstanceKey(args[1])
			locationPath := args[2]
			content, err := pkg.GenerateContent(stateStore, *locationConfig, locationKey, instanceKey, locationPath)
			if err != nil {
				logger.Printf("failed to generate content: %s", err)
				return err
			}

			fmt.Print(content)

			return nil
		},
		Args: cobra.ExactArgs(3),
	}

	var runUpdates = &cobra.Command{
		Use:   "update",
		Short: "run update of operatons for a location and instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger, err := setupLogger()
			if err != nil {
				fmt.Println("failed to setup logger: %s", err)
				return err
			}

			locationKey := pkg.LocationKey(args[0])
			locationConfig, stateStore, config, err := setup(locationKey)
			if err != nil {
				logger.Printf("failed to setup: %s", err)
				return err
			}

			instanceKey := pkg.InstanceKey(args[1])
			locationPath := args[2]
			err = pkg.Update(stateStore, *locationConfig, locationKey, instanceKey, locationPath)
			if err != nil {
				logger.Printf("failed to update: %s", err)
				return err
			}

			err = runPostCommands(config, logger)
			if err != nil {
				logger.Printf("failed to run post commands: %s", err)
				return err
			}

			return nil
		},
		Args: cobra.ExactArgs(3),
	}

	var startUpdate = &cobra.Command{
		Use:   "start-update",
		Short: "start update of operatons for a location and instance in the background i.e. calls update and then returns",
		RunE: func(cmd *cobra.Command, args []string) error {
			locationKey := pkg.LocationKey(args[0])
			instanceKey := pkg.InstanceKey(args[1])
			locationPath := args[2]

			executable, err := os.Executable()
			if err != nil {
				return err
			}

			updateCommand := exec.Command(executable, "update", string(locationKey), string(instanceKey), locationPath)
			err = updateCommand.Start()
			if err != nil {
				return err
			}

			return nil
		},
		Args: cobra.ExactArgs(3),
	}

	// var printDefaults = &cobra.Command{
	// 	Use:   "print-defaults",
	// 	Short: "print the default config",
	// 	Run: func(cmd *cobra.Command, args []string) {

	// 	},
	// }

	var setState = &cobra.Command{
		Use:   "set-state",
		Short: "set state for an operation",
		RunE: func(cmd *cobra.Command, args []string) error {
			logger, err := setupLogger()
			if err != nil {
				fmt.Println("failed to setup logger: %s", err)
				return err
			}

			locationKey := pkg.LocationKey(args[0])
			_, stateStore, config, err := setup(locationKey)
			if err != nil {
				logger.Printf("failed to setup: %s", err)
				return err
			}

			instanceKey := pkg.InstanceKey(args[1])
			operationName := pkg.OperationName(args[2])
			err = stateStore.Set(locationKey, instanceKey, operationName, args[3])
			if err != nil {
				logger.Printf("failed to set state: %s", err)
				return err
			}

			err = runPostCommands(config, logger)
			if err != nil {
				logger.Printf("failed to run post commands: %s", err)
				return err
			}

			return nil
		},
		Args: cobra.ExactArgs(4),
	}

	rootCmd.AddCommand(runUpdates)
	// rootCmd.AddCommand(printDefaults)
	rootCmd.AddCommand(generateCmd)
	rootCmd.AddCommand(setState)
	rootCmd.AddCommand(startUpdate)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
