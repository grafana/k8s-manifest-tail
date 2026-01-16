package cmd

import (
	"fmt"
	"github.com/spf13/cobra"
)

var listCmd = &cobra.Command{
	Use:     "list",
	Short:   "List the Kubernetes objects matched by the configuration",
	PreRunE: LoadConfiguration,
	RunE:    runList,
}

func init() {
	rootCmd.AddCommand(listCmd)
}

func runList(cmd *cobra.Command, args []string) error {
	return fmt.Errorf("list command not implemented yet")
}
