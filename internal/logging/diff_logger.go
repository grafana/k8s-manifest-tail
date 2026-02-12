package logging

import (
	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/manifest"
	"go.opentelemetry.io/otel/log"
)

// DiffLogger emits information about manifest diffs.
type DiffLogger interface {
	Log(diff *manifest.Diff)
}

// NewDiffLogger returns the appropriate logger for the configured mode.
func NewDiffLogger(cfg config.LoggingConfig, logger log.Logger) DiffLogger {
	switch cfg.LogDiffs {
	case config.LogDiffsCompact:
		return &CompactDiffLogger{logger: logger}
	case config.LogDiffsDetailed:
		return &DetailedDiffLogger{logger: logger}
	default:
		return &NullLogger{}
	}
}

// NewManifestLogger returns the appropriate logger for the configured mode.
func NewManifestLogger(cfg config.LoggingConfig, logger log.Logger) DiffLogger {
	if cfg.LogManifests {
		return &ManifestLogger{logger: logger}
	}
	return &NullLogger{}
}
