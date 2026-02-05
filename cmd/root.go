package cmd

import (
	"github.com/spf13/cobra"
	"io"
)

var (
	configPath                string
	kubeconfigOverride        string
	outputDirOverride         string
	outputFormatOverride      string
	refreshIntervalOverride   string
	namespacesOverride        []string
	excludeNamespacesOverride []string

	rootCmd = &cobra.Command{
		Use:   "k8s-manifest-tail",
		Short: "Kubernetes manifest fetcher",
	}
)

// Execute runs the CLI with the default arguments.
func Execute() error {
	return rootCmd.Execute()
}

// ExecuteWithArgs executes the CLI with the provided arguments and IO writers.
func ExecuteWithArgs(args []string, stdout, stderr io.Writer) error {
	rootCmd.SetArgs(args)
	rootCmd.SetOut(stdout)
	rootCmd.SetErr(stderr)
	return rootCmd.Execute()
}

func init() {
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config.yaml", "Path to configuration file")
	rootCmd.PersistentFlags().StringVar(&kubeconfigOverride, "kubeconfig", "", "Path to kubeconfig (overrides config file)")
	rootCmd.PersistentFlags().StringVarP(&outputDirOverride, "output-directory", "o", "", "Directory for manifest output (overrides config file)")
	rootCmd.PersistentFlags().StringVarP(&outputFormatOverride, "output-format", "f", "", "Output format: yaml or json (overrides config file)")
	rootCmd.PersistentFlags().StringVar(&refreshIntervalOverride, "refresh-interval", "", "Interval for full refresh (overrides config file)")
	rootCmd.PersistentFlags().StringSliceVarP(&namespacesOverride, "namespaces", "n", nil, "Namespaces to include globally (overrides config file)")
	rootCmd.PersistentFlags().StringSliceVar(&excludeNamespacesOverride, "exclude-namespaces", nil, "Namespaces to exclude globally (overrides config file)")
}
