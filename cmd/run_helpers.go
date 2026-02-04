package cmd

import (
	"github.com/grafana/k8s-manifest-tail/internal/manifest"
)

var manifestProcessor manifest.Processor

func GetManifestProcessor() manifest.Processor {
	if manifestProcessor == nil {
		SetManifestProcessor(manifest.NewProcessor(Configuration))
	}
	return manifestProcessor
}

// SetManifestProcessor overrides the manifest processor used by the run command (primarily for tests).
func SetManifestProcessor(p manifest.Processor) {
	manifestProcessor = p
}
