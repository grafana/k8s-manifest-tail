package logging

import "github.com/grafana/k8s-manifest-tail/internal/manifest"

// NullLogger discards all diff events.
type NullLogger struct{}

// Log implements DiffLogger.
func (NullLogger) Log(*manifest.Diff) {}
