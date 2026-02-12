package logging

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/grafana/k8s-manifest-tail/internal/manifest"
	"github.com/grafana/k8s-manifest-tail/internal/telemetry"
	"go.opentelemetry.io/otel/log"
	"strings"
)

type ManifestLogger struct {
	logger log.Logger
}

func (l *ManifestLogger) Log(diff *manifest.Diff) {
	if diff == nil || diff.Current == nil {
		return
	}
	var manifestPayload bytes.Buffer
	rawJSON, err := diff.Current.MarshalJSON()
	if err != nil {
		return
	}
	err = json.Compact(&manifestPayload, rawJSON)
	if err != nil {
		return
	}

	if diff.Current.GetNamespace() == "" {
		telemetry.Info(
			l.logger,
			manifestPayload.String(),
			log.String("action", "manifest"),
			log.String(fmt.Sprintf("k8s.%s.name", strings.ToLower(diff.Current.GetKind())), diff.Current.GetName()),
		)
	} else {
		telemetry.Info(
			l.logger,
			manifestPayload.String(),
			log.String("action", "manifest"),
			log.String(fmt.Sprintf("k8s.%s.name", strings.ToLower(diff.Current.GetKind())), diff.Current.GetName()),
			log.String("k8s.namespace.name", diff.Current.GetNamespace()),
		)
	}
}
