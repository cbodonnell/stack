package commands

import (
	"fmt"

	"github.com/cbodonnell/stack/pkg/config"
	"github.com/cbodonnell/stack/pkg/run"
	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:           "run",
	Short:         "Run the development environment",
	SilenceUsage:  true,
	SilenceErrors: false,
	RunE: func(cmd *cobra.Command, args []string) error {
		return Run()
	},
}

var configFilePath string

func init() {
	runCmd.Flags().StringVarP(&configFilePath, "config-file", "f", "./stack.yaml", "path to the stack config file")
	rootCmd.AddCommand(runCmd)
}

func Run() error {
	cfg, err := config.LoadConfig(configFilePath)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}
	return run.Run(cfg)
}
