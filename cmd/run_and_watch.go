package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var runAndWatchCmd = &cobra.Command{
	Use:     "run-and-watch",
	Short:   "Fetch manifests and keep watching for changes",
	PreRunE: LoadConfiguration,
	RunE:    runRunAndWatch,
}

func init() {
	rootCmd.AddCommand(runAndWatchCmd)
}

func runRunAndWatch(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("run and watch command not implemented yet")
}
