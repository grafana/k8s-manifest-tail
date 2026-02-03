package logging

import (
	"fmt"
	"github.com/grafana/k8s-manifest-tail/internal/manifest"
	"github.com/grafana/k8s-manifest-tail/internal/telemetry"
	"go.opentelemetry.io/otel/log"
	"strings"
)

// DetailedDiffLogger prints a detailed diff statement
type DetailedDiffLogger struct {
	logger log.Logger
}

// Log implements DiffLogger.
func (l *DetailedDiffLogger) Log(diff *manifest.Diff) {
	if l.logger == nil || diff == nil || diff.Current == nil {
		return
	}
	target := diff.Current
	if target.GetNamespace() == "" {
		telemetry.Info(
			l.logger,
			fmt.Sprintf("Object changed: %s %s", target.GetKind(), target.GetName()),
			log.String(fmt.Sprintf("k8s.%s.name", strings.ToLower(target.GetKind())), target.GetName()),
		)
	} else {
		telemetry.Info(
			l.logger,
			fmt.Sprintf("Object changed: %s %s/%s", target.GetKind(), target.GetNamespace(), target.GetName()),
			log.String(fmt.Sprintf("k8s.%s.name", strings.ToLower(target.GetKind())), target.GetName()),
			log.String("k8s.namespace.name", target.GetNamespace()),
		)
	}
}
