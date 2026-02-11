package cmd

import (
	"fmt"
	"github.com/grafana/k8s-manifest-tail/internal"
	"strings"

	"github.com/spf13/cobra"
)

var (
	versionCmd = &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		RunE: func(cmd *cobra.Command, args []string) error {
			out := internal.Version
			if internal.GitCommit != "" {
				commit := internal.GitCommit
				if len(commit) > 7 {
					commit = commit[:7]
				}
				out = fmt.Sprintf("%s (%s)", internal.Version, commit)
			}
			if strings.TrimSpace(out) == "" {
				out = "dev"
			}
			_, _ = fmt.Fprintln(cmd.OutOrStdout(), out)
			return nil
		},
	}
)

func init() {
	rootCmd.AddCommand(versionCmd)
}
