package logging

import "github.com/grafana/k8s-manifest-tail/internal/manifest"

// NullDiffLogger discards all diff events.
type NullDiffLogger struct{}

// Log implements DiffLogger.
func (NullDiffLogger) Log(*manifest.Diff) {}
