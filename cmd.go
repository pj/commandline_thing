package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/pj/commandline_thing/pkg"
	"github.com/spf13/cobra"
)

func setup(locationKey pkg.LocationKey) (*pkg.Location, pkg.StateStore, error) {
	availableOperations := pkg.LoadAvailableOperations()

	allLocations, err := pkg.LoadConfig(availableOperations)

	if err != nil {
		return nil, nil, err
	}

	config, ok := allLocations.Configs[locationKey]
	if !ok {
		return nil, nil, fmt.Errorf("config not found: %s", locationKey)
	}

	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, nil, fmt.Errorf("failed to get user's home directory: %w", err)
	}

	stateDBPath := filepath.Join(homeDir, ".config", "commandline_thing", "state.db")

	state, err := pkg.NewSQLiteState(stateDBPath)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create state: %w", err)
	}

	return &config, state, nil
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
			locationKey := pkg.LocationKey(args[0])
			config, stateStore, err := setup(locationKey)
			if err != nil {
				return err
			}

			instanceKey := pkg.InstanceKey(args[1])
			content, err := pkg.GenerateContent(stateStore, *config, locationKey, instanceKey)
			if err != nil {
				return err
			}

			fmt.Print(content)

			return nil
		},
		Args: cobra.ExactArgs(2),
	}

	var runUpdates = &cobra.Command{
		Use:   "update",
		Short: "run update of operatons for a location and instance",
		RunE: func(cmd *cobra.Command, args []string) error {
			locationKey := pkg.LocationKey(args[0])
			config, stateStore, err := setup(locationKey)
			if err != nil {
				return err
			}

			instanceKey := pkg.InstanceKey(args[1])

			err = pkg.Update(stateStore, *config, locationKey, instanceKey)
			if err != nil {
				return err
			}

			return nil
		},
		Args: cobra.ExactArgs(2),
	}

	var startUpdate = &cobra.Command{
		Use:   "start-update",
		Short: "start update of operatons for a location and instance in the background i.e. calls update and then returns",
		RunE: func(cmd *cobra.Command, args []string) error {
			locationKey := pkg.LocationKey(args[0])
			instanceKey := pkg.InstanceKey(args[1])

			executable, err := os.Executable()
			if err != nil {
				return err
			}

			updateCommand := exec.Command(executable, "update", string(locationKey), string(instanceKey))
			err = updateCommand.Start()
			if err != nil {
				return err
			}

			return nil
		},
		Args: cobra.ExactArgs(2),
	}

	// var printDefaults = &cobra.Command{
	// 	Use:   "print-defaults",
	// 	Short: "print the default config",
	// 	Run: func(cmd *cobra.Command, args []string) {

	// 	},
	// }

	// var setState = &cobra.Command{
	// 	Use:   "set-state",
	// 	Short: "set state for an operation",
	// 	RunE: func(cmd *cobra.Command, args []string) error {
	// 		locationKey := pkg.LocationKey(args[0])
	// 		locationConfig, state, err := setup(locationKey)
	// 		if err != nil {
	// 			return err
	// 		}

	// 		instanceKey := pkg.InstanceKey(args[1])

	// 		operationName := pkg.OperationName(args[2])
	// 		for _, operationWrapper := range locationConfig.Operations {
	// 			if operationWrapper.Operation.Name() == operationName {
	// 				state, err := state.Get(locationKey, instanceKey, operationName)
	// 				if err != nil {
	// 					return err
	// 				}
	// 				operationWrapper.Operation.Update(state)

	// 				return nil
	// 			}
	// 		}
	// 		return fmt.Errorf("operation not found: %s", operationName)
	// 	},
	// 	Args: cobra.ExactArgs(3),
	// }

	rootCmd.AddCommand(runUpdates)
	// rootCmd.AddCommand(printDefaults)
	rootCmd.AddCommand(generateCmd)
	// rootCmd.AddCommand(setState)
	rootCmd.AddCommand(startUpdate)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
