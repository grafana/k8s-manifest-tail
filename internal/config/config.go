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

	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return &cfg, nil
}

// Validate ensures the configuration is internally consistent.
func (cfg *Config) Validate() error {
	for i, rule := range cfg.Objects {
		if err := rule.Validate(); err != nil {
			return fmt.Errorf("validate object rule %d: %w", i+1, err)
		}
	}
	err := checkForDuplicates(cfg.Namespaces)
	if err != nil {
		return fmt.Errorf("global inclusion namespaces has duplicate: %w", err)
	}
	err = checkForDuplicates(cfg.ExcludeNamespaces)
	if err != nil {
		return fmt.Errorf("global exclusion namespaces has duplicate: %w", err)
	}
	return nil
}

// Validate ensures an object rule is internally consistent.
func (rule *ObjectRule) Validate() error {
	return checkForDuplicates(rule.Namespaces)
}

func checkForDuplicates(namespaces []string) error {
	seen := make(map[string]struct{}, len(namespaces))
	for _, ns := range namespaces {
		if ns == "" {
			continue
		}
		if _, ok := seen[ns]; ok {
			return fmt.Errorf("duplicate namespace %s", ns)
		}
		seen[ns] = struct{}{}
	}
	return nil
}
