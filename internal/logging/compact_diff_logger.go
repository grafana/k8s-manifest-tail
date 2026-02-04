package logging

import (
	"fmt"
	"github.com/grafana/k8s-manifest-tail/internal/manifest"
	"github.com/grafana/k8s-manifest-tail/internal/telemetry"
	"go.opentelemetry.io/otel/log"
	"strings"
)

// CompactDiffLogger prints a compact diff statement.
type CompactDiffLogger struct {
	logger log.Logger
}

// Log implements DiffLogger.
func (l *CompactDiffLogger) Log(diff *manifest.Diff) {
	if l.logger == nil || diff == nil || (diff.Previous == nil && diff.Current == nil) {
		return
	}

	target := diff.Current
	action := "modified"
	if diff.Previous == nil {
		action = "created"
	} else if target == nil {
		target = diff.Previous
		action = "deleted"
	}

	if target.GetNamespace() == "" {
		telemetry.Info(
			l.logger,
			fmt.Sprintf("Object %s: %s %s", action, target.GetKind(), target.GetName()),
			log.String(fmt.Sprintf("k8s.%s.name", strings.ToLower(target.GetKind())), target.GetName()),
		)
	} else {
		telemetry.Info(
			l.logger,
			fmt.Sprintf("Object %s: %s %s/%s", action, target.GetKind(), target.GetNamespace(), target.GetName()),
			log.String(fmt.Sprintf("k8s.%s.name", strings.ToLower(target.GetKind())), target.GetName()),
			log.String("k8s.namespace.name", target.GetNamespace()),
		)
	}
}
