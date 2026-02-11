package cmd

import (
	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/manifest"
	"go.opentelemetry.io/otel/log"
)

var manifestProcessor manifest.Processor

func GetManifestProcessor(cfg *config.Config, logger log.Logger) manifest.Processor {
	if manifestProcessor == nil {
		var writer manifest.Processor = manifest.NewWriter(cfg.Output)

		if cfg.Logging.LogManifests {
			writer = manifest.NewLogger(logger, writer)
		}
		manifestProcessor = manifest.NewFilterProcessor(
			writer,
			manifest.RemoveStatusFilter{},
			manifest.RemoveMetadataFieldsFilter{},
			manifest.RedactEnvValuesFilter{},
		)
	}

	return manifestProcessor
}

// SetManifestProcessor overrides the manifest processor used by the run command (primarily for tests).
func SetManifestProcessor(p manifest.Processor) {
	manifestProcessor = p
}
