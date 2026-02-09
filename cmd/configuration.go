package cmd

import (
	"fmt"
	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/spf13/cobra"
	"strings"
)

var Configuration *config.Config

// LoadConfiguration reads the config file and applies flag overrides.
func LoadConfiguration(cmd *cobra.Command, args []string) error {
	cfg, err := config.Load(configPath)
	if err != nil {
		return err
	}

	config.ApplyEnvOverrides(cfg)

	if kubeconfigOverride != "" {
		cfg.Kubeconfig = kubeconfigOverride
	}

	// Output directory
	if cfg.Output.Directory == "" {
		cfg.Output.Directory = "output"
	}
	if outputDirOverride != "" {
		cfg.Output.Directory = outputDirOverride
	}

	// Output manifest file format
	if cfg.Output.Format == "" {
		cfg.Output.Format = config.OutputFormatYAML
	}
	if outputFormatOverride != "" {
		cfg.Output.Format = config.OutputFormat(strings.ToLower(outputFormatOverride))
	}
	err = validateOutputFormat(cfg.Output.Format)
	if err != nil {
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

	if len(cfg.Objects) == 0 {
		return fmt.Errorf("no objects found in configuration file")
	}

	Configuration = cfg
	return Configuration.Validate()
}

func validateOutputFormat(format config.OutputFormat) error {
	switch format {
	case config.OutputFormatYAML, config.OutputFormatJSON:
		return nil
	default:
		return fmt.Errorf("invalid output format %q (expected yaml or json)", format)
	}
}

// ResetConfiguration clears the cached configuration (useful for tests).
func ResetConfiguration() {
	Configuration = nil
}
