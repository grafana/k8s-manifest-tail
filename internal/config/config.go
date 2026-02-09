package config

import (
	"fmt"
	"os"
	"regexp"
	"strings"
	"time"

	"gopkg.in/yaml.v3"
)

// OutputFormat enumerates the supported serialization formats.
type OutputFormat string

const (
	DefaultRefreshInterval string       = "24h"
	OutputFormatYAML       OutputFormat = "yaml"
	OutputFormatJSON       OutputFormat = "json"
)

// OutputConfig controls how manifests are written.
type OutputConfig struct {
	Directory string       `mapstructure:"directory" yaml:"directory"`
	Format    OutputFormat `mapstructure:"format" yaml:"format"`
}

// ObjectRule describes which Kubernetes objects to collect.
type ObjectRule struct {
	APIVersion  string   `mapstructure:"apiVersion" yaml:"apiVersion"`
	Kind        string   `mapstructure:"kind" yaml:"kind"`
	Namespaces  []string `mapstructure:"namespaces" yaml:"namespaces"`
	NamePattern string   `mapstructure:"namePattern" yaml:"namePattern"`
}

// Config captures all supported configuration settings.
type Config struct {
	Output                  OutputConfig  `mapstructure:"output" yaml:"output"`
	Logging                 LoggingConfig `mapstructure:"logging" yaml:"logging"`
	RefreshInterval         string        `mapstructure:"refreshInterval" yaml:"refreshInterval"`
	RefreshIntervalDuration time.Duration
	Namespaces              []string     `mapstructure:"namespaces" yaml:"namespaces"`
	ExcludeNamespaces       []string     `mapstructure:"excludeNamespaces" yaml:"excludeNamespaces"`
	Objects                 []ObjectRule `mapstructure:"objects" yaml:"objects"`
	KubeconfigPath          string       `yaml:"-" mapstructure:"-"`
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

// ApplyEnvOverrides applies environment variable overrides for selected fields.
func ApplyEnvOverrides(cfg *Config) {
	if cfg == nil {
		return
	}
	if value := strings.TrimSpace(os.Getenv("K8S_MANIFEST_TAIL_OUTPUT_DIRECTORY")); value != "" {
		cfg.Output.Directory = value
	}
	if value := strings.TrimSpace(os.Getenv("K8S_MANIFEST_TAIL_OUTPUT_FORMAT")); value != "" {
		cfg.Output.Format = OutputFormat(strings.ToLower(value))
	}
	if value := strings.TrimSpace(os.Getenv("K8S_MANIFEST_TAIL_LOGGING_LOG_DIFFS")); value != "" {
		cfg.Logging.LogDiffs = LogDiffMode(strings.ToLower(value))
	}
	if value := strings.TrimSpace(os.Getenv("K8S_MANIFEST_TAIL_REFRESH_INTERVAL")); value != "" {
		cfg.RefreshInterval = value
	}
	if value := strings.TrimSpace(os.Getenv("K8S_MANIFEST_TAIL_NAMESPACES")); value != "" {
		cfg.Namespaces = strings.Split(value, ",")
	}
	if value := strings.TrimSpace(os.Getenv("K8S_MANIFEST_TAIL_EXCLUDE_NAMESPACES")); value != "" {
		cfg.ExcludeNamespaces = strings.Split(value, ",")
	}
}

// Validate ensures the configuration is internally consistent.
func (cfg *Config) Validate() error {
	if err := cfg.Logging.Validate(); err != nil {
		return fmt.Errorf("validate logging config: %w", err)
	}
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
	_, err = cfg.GetRefreshInterval()
	if err != nil {
		return err
	}
	return nil
}

func (cfg *Config) GetRefreshInterval() (time.Duration, error) {
	interval := cfg.RefreshInterval
	if strings.TrimSpace(interval) == "" {
		interval = DefaultRefreshInterval
	}
	duration, err := time.ParseDuration(interval)
	if err != nil {
		return 0, fmt.Errorf("invalid refresh interval %q: %w", interval, err)
	}
	return duration, nil
}

// LoggingConfig controls optional logging behavior.
type LoggingConfig struct {
	LogDiffs LogDiffMode `mapstructure:"logDiffs" yaml:"logDiffs"`
	OTLP     OTLPConfig  `mapstructure:"otlp" yaml:"otlp"`
}

// Mode returns the normalized diff logging mode.
func (l LoggingConfig) Mode() LogDiffMode {
	if l.LogDiffs == "" {
		return LogDiffsDisabled
	}
	return l.LogDiffs
}

// Validate ensures logging settings are valid.
func (l LoggingConfig) Validate() error {
	switch mode := l.Mode(); mode {
	case LogDiffsDisabled, LogDiffsCompact, LogDiffsDetailed:
	default:
		return fmt.Errorf("unsupported diff logging mode %q", mode)
	}
	if err := l.OTLP.Validate(); err != nil {
		return fmt.Errorf("validate otlp logging config: %w", err)
	}
	return nil
}

// LogDiffMode enumerates supported diff logging modes.
type LogDiffMode string

const (
	LogDiffsDisabled LogDiffMode = "disabled"
	LogDiffsCompact  LogDiffMode = "compact"
	LogDiffsDetailed LogDiffMode = "detailed"
)

// UnmarshalYAML allows the mode to be provided as a string or bool.
func (m *LogDiffMode) UnmarshalYAML(value *yaml.Node) error {
	if value == nil {
		*m = LogDiffsDisabled
		return nil
	}

	var asString string
	if err := value.Decode(&asString); err == nil {
		normalized := strings.ToLower(strings.TrimSpace(asString))
		switch normalized {
		case "", "false", "disabled":
			*m = LogDiffsDisabled
		case "compact":
			*m = LogDiffsCompact
		case "detailed":
			*m = LogDiffsDetailed
		default:
			*m = LogDiffMode(normalized)
		}
		return nil
	}

	var asBool bool
	if err := value.Decode(&asBool); err == nil {
		if asBool {
			*m = LogDiffsDetailed
		} else {
			*m = LogDiffsDisabled
		}
		return nil
	}

	return fmt.Errorf("invalid logDiffs value")
}

// OTLPConfig controls OTLP logging export.
type OTLPConfig struct {
	Enabled  bool   `mapstructure:"enabled" yaml:"enabled"`
	Endpoint string `mapstructure:"endpoint" yaml:"endpoint"`
	Insecure bool   `mapstructure:"insecure" yaml:"insecure"`
}

// ShouldEnable returns true when OTLP logging should be initialized.
func (c OTLPConfig) ShouldEnable() bool {
	return c.Enabled || c.Endpoint != ""
}

// Validate ensures OTLP config is coherent.
func (c OTLPConfig) Validate() error {
	return nil
}

// Validate ensures an object rule is internally consistent.
func (rule *ObjectRule) Validate() error {
	if err := checkForDuplicates(rule.Namespaces); err != nil {
		return err
	}
	if strings.TrimSpace(rule.NamePattern) != "" {
		if _, err := regexp.Compile(rule.NamePattern); err != nil {
			return fmt.Errorf("invalid namePattern %q: %w", rule.NamePattern, err)
		}
	}
	return nil
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
