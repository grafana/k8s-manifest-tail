package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
)

var (
	version    = "dev"
	gitCommit  = ""
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := version
			if gitCommit != "" {
				commit := gitCommit
				if len(commit) > 7 {
					commit = commit[:7]
				}
				out = fmt.Sprintf("%s (%s)", version, commit)
			}
			if strings.TrimSpace(out) == "" {
				out = "dev"
			}
			fmt.Fprintln(cmd.OutOrStdout(), out)
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
