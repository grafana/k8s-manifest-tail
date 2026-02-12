package cmd

import (
	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/manifest"
)

var manifestProcessor manifest.Processor

func GetManifestProcessor(cfg *config.Config) manifest.Processor {
	if manifestProcessor == nil {
		manifestProcessor = manifest.NewFilterProcessor(
			manifest.NewWriter(cfg.Output),
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
