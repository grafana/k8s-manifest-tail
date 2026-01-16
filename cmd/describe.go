package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

var describeCmd = &cobra.Command{
	Use:     "describe",
	Short:   "Describe the configuration",
	PreRunE: LoadConfiguration,
	RunE:    runDescribe,
}

func init() {
	rootCmd.AddCommand(describeCmd)
}

func runDescribe(cmd *cobra.Command, args []string) error {
	_, _ = fmt.Fprintln(cmd.OutOrStdout(), Configuration.Describe())
	return nil
}
