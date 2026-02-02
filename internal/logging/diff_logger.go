package logging

import (
	"fmt"
	"io"

	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/manifest"
)

// DiffLogger emits information about manifest diffs.
type DiffLogger interface {
	Log(diff *manifest.Diff)
}

// NewDiffLogger returns the appropriate logger for the configured mode.
func NewDiffLogger(cfg config.LoggingConfig, out io.Writer) DiffLogger {
	switch cfg.Mode() {
	case config.LogDiffsCompact:
		return &CompactDiffLogger{out: out}
	case config.LogDiffsDetailed:
		return &DetailedDiffLogger{out: out}
	default:
		return NullDiffLogger{}
	}
}

// NullDiffLogger discards all diff events.
type NullDiffLogger struct{}

// Log implements DiffLogger.
func (NullDiffLogger) Log(_ *manifest.Diff) {}

// CompactDiffLogger emits terse change notifications.
type CompactDiffLogger struct {
	out io.Writer
}

// Log implements DiffLogger.
func (l *CompactDiffLogger) Log(diff *manifest.Diff) {
	if diff == nil || l.out == nil {
		return
	}
	_, _ = fmt.Fprintf(l.out, "Object changed: %s %s/%s\n", diff.Current.GetKind(), namespaceOrDash(diff.Current.GetNamespace()), diff.Current.GetName())
}

// DetailedDiffLogger is a placeholder for future detailed diff output.
type DetailedDiffLogger struct {
	out io.Writer
}

// Log implements DiffLogger.
func (l *DetailedDiffLogger) Log(diff *manifest.Diff) {
	if diff == nil {
		return
	}
	// Fallback to the compact message until detailed output is implemented.
	_, _ = fmt.Fprintf(l.out, "Object changed: %s %s/%s\n", diff.Current.GetKind(), namespaceOrDash(diff.Current.GetNamespace()), diff.Current.GetName())
}

func namespaceOrDash(ns string) string {
	if ns == "" {
		return "-"
	}
	return ns
}
