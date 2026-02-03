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
	if cfg.Mode() == config.LogDiffsCompact {
		return &CompactDiffLogger{logger: logger}
	} else if cfg.Mode() == config.LogDiffsDetailed {
		return &CompactDiffLogger{logger: logger}
	}
	return &NullDiffLogger{}
}
