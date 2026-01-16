package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runCmd = &cobra.Command{
	Use:     "run",
	Short:   "Fetch manifests once and exit",
	PreRunE: LoadConfiguration,
	RunE:    runRun,
}

func init() {
	rootCmd.AddCommand(runCmd)
}

func runRun(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("run command not implemented yet")
}
