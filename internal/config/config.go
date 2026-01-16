package config

import (
	"fmt"
	"os"

	"gopkg.in/yaml.v3"
)

// OutputFormat enumerates the supported serialization formats.
type OutputFormat string

const (
	OutputFormatYAML OutputFormat = "yaml"
	OutputFormatJSON OutputFormat = "json"
)

// OutputConfig controls how manifests are written.
type OutputConfig struct {
	Directory string       `mapstructure:"directory" yaml:"directory"`
	Format    OutputFormat `mapstructure:"format" yaml:"format"`
}

// ObjectRule describes which Kubernetes objects to collect.
type ObjectRule struct {
	APIVersion string   `mapstructure:"apiVersion" yaml:"apiVersion"`
	Kind       string   `mapstructure:"kind" yaml:"kind"`
	Namespaces []string `mapstructure:"namespaces" yaml:"namespaces"`
}

// Config captures all supported configuration settings.
type Config struct {
	Kubeconfig        string       `mapstructure:"kubeconfig" yaml:"kubeconfig"`
	Output            OutputConfig `mapstructure:"output" yaml:"output"`
	RefreshInterval   string       `mapstructure:"refreshInterval" yaml:"refreshInterval"`
	Namespaces        []string     `mapstructure:"namespaces" yaml:"namespaces"`
	ExcludeNamespaces []string     `mapstructure:"excludeNamespaces" yaml:"excludeNamespaces"`
	Objects           []ObjectRule `mapstructure:"objects" yaml:"objects"`
}

// Load reads configuration data from the supplied file path.
func Load(path string) (*Config, error) {
	if path == "" {
		return nil, fmt.Errorf("config path is required")
	}

	file, err := os.Open(path)
	if err != nil {
		return nil, fmt.Errorf("open config: %w", err)
	}
	defer func(file *os.File) { _ = file.Close() }(file)

	decoder := yaml.NewDecoder(file)
	decoder.KnownFields(true)

	var cfg Config
	if err := decoder.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("decode config: %w", err)
	}

	return &cfg, nil
}
