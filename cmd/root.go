package cmd

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"github.com/grafana/k8s-manifest-tail/internal/config"
)

var (
	Configuration             *config.Config
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

// Execute runs the CLI.
func Execute() error {
	return rootCmd.Execute()
}

func init() {
	Configuration = nil
	rootCmd.PersistentFlags().StringVarP(&configPath, "config", "c", "config.yaml", "Path to configuration file")
	rootCmd.PersistentFlags().StringVar(&kubeconfigOverride, "kubeconfig", "", "Path to kubeconfig (overrides config file)")
	rootCmd.PersistentFlags().StringVarP(&outputDirOverride, "output-directory", "o", "", "Directory for manifest output (overrides config file)")
	rootCmd.PersistentFlags().StringVarP(&outputFormatOverride, "output-format", "f", "", "Output format: yaml or json (overrides config file)")
	rootCmd.PersistentFlags().StringVar(&refreshIntervalOverride, "refresh-interval", "", "Interval for full refresh (overrides config file)")
	rootCmd.PersistentFlags().StringSliceVarP(&namespacesOverride, "namespaces", "n", nil, "Namespaces to include globally (overrides config file)")
	rootCmd.PersistentFlags().StringSliceVar(&excludeNamespacesOverride, "exclude-namespaces", nil, "Namespaces to exclude globally (overrides config file)")
}

func LoadConfiguration(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	if cfg.Output.Directory == "" {
		cfg.Output.Directory = "output"
	}
	if cfg.Output.Format == "" {
		cfg.Output.Format = config.OutputFormatYAML
	}

	if kubeconfigOverride != "" {
		cfg.Kubeconfig = kubeconfigOverride
	}
	if outputDirOverride != "" {
		cfg.Output.Directory = outputDirOverride
	}
	if outputFormatOverride != "" {
		format := config.OutputFormat(strings.ToLower(outputFormatOverride))
		if err := validateOutputFormat(format); err != nil {
			return err
		}
		cfg.Output.Format = format
	} else if err := validateOutputFormat(cfg.Output.Format); err != nil {
		return err
	}
	if refreshIntervalOverride != "" {
		cfg.RefreshInterval = refreshIntervalOverride
	}
	if len(namespacesOverride) > 0 {
		cfg.Namespaces = namespacesOverride
	}
	if len(excludeNamespacesOverride) > 0 {
		cfg.ExcludeNamespaces = excludeNamespacesOverride
	}

	Configuration = cfg
	return RequireObjects(cmd, args)
}

func RequireObjects(cmd *cobra.Command, args []string) error {
	if Configuration != nil && len(Configuration.Objects) == 0 {
		return fmt.Errorf("no objects found in configuration file")
	}
	return nil
}

func validateOutputFormat(format config.OutputFormat) error {
	switch format {
	case config.OutputFormatYAML, config.OutputFormatJSON:
		return nil
	default:
		return fmt.Errorf("invalid output format %q (expected yaml or json)", format)
	}
}
