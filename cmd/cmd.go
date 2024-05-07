package cmd

import (
	"fmt"

	"github.com/pj/commandline_thing/pkg"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "commandline_thing",
	Short: "Generate various things related to tmux",
	Long:  `Generate various things related to tmux`,
}

var generatePaneLeft = &cobra.Command{
	Use:   "generate-pane-left",
	Short: "Generate pane content",
	Long:  `This subcommand does something specific.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		availableOperations := pkg.LoadAvailableOperations()
		config, err := pkg.LoadConfig(availableOperations)
		if err != nil {
			return err
		}

		output, err := config.GeneratePaneLeft()
		if err != nil {

		}
		fmt.Print(output)

		return nil
	},
}

var generatePaneRight = &cobra.Command{
	Use:   "generate-pane-right",
	Short: "Generate pane content",
	Long:  `This subcommand does something specific.`,
	RunE: func(cmd *cobra.Command, args []string) error {
		availableOperations := pkg.LoadAvailableOperations()
		config, err := pkg.LoadConfig(availableOperations)
		if err != nil {
			return err
		}

		output, err := config.GeneratePaneLeft()
		if err != nil {

		}
		fmt.Print(output)

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
	rootCmd.AddCommand(generatePaneLeft)
	rootCmd.AddCommand(generatePaneRight)
	rootCmd.AddCommand(runUpdates)
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
	}
}
