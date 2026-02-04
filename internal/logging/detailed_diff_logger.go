package logging

import (
	"encoding/json"
	"fmt"
	"github.com/grafana/k8s-manifest-tail/internal/telemetry"
	"strings"

	"github.com/grafana/k8s-manifest-tail/internal/manifest"
	"go.opentelemetry.io/otel/log"
)

// DetailedDiffLogger prints a JSON diff statement.
type DetailedDiffLogger struct {
	logger log.Logger
}

type DetailedDiffReport struct {
	Kind      string      `json:"kind"`
	Name      string      `json:"name"`
	Namespace string      `json:"namespace,omitempty"`
	Action    string      `json:"action"`
	Previous  interface{} `json:"previous,omitempty"`
	Current   interface{} `json:"current,omitempty"`
}

// Log implements DiffLogger.
func (l *DetailedDiffLogger) Log(diff *manifest.Diff) {
	if l.logger == nil || diff == nil || (diff.Previous == nil && diff.Current == nil) {
		return
	}

	payload := &DetailedDiffReport{}
	if diff.Previous == nil {
		payload.Kind = diff.Current.GetKind()
		payload.Name = diff.Current.GetName()
		payload.Namespace = diff.Current.GetNamespace()
		payload.Action = "created"
	} else if diff.Current == nil {
		payload.Kind = diff.Previous.GetKind()
		payload.Name = diff.Previous.GetName()
		payload.Namespace = diff.Previous.GetNamespace()
		payload.Action = "deleted"
	} else {
		payload.Kind = diff.Current.GetKind()
		payload.Name = diff.Current.GetName()
		payload.Namespace = diff.Current.GetNamespace()
		payload.Action = "modified"
		previousDiff, currentDiff := GetMinimalDifference(diff)
		payload.Previous = previousDiff
		payload.Current = currentDiff
	}

	jsonBytes, err := json.Marshal(payload)
	if err != nil {
		return
	}

	if payload.Namespace == "" {
		telemetry.Info(
			l.logger,
			string(jsonBytes),
			log.String(fmt.Sprintf("k8s.%s.name", strings.ToLower(payload.Kind)), payload.Name),
		)
	} else {
		telemetry.Info(
			l.logger,
			string(jsonBytes),
			log.String(fmt.Sprintf("k8s.%s.name", strings.ToLower(payload.Kind)), payload.Name),
			log.String("k8s.namespace.name", payload.Namespace),
		)
	}
}

func GetMinimalDifference(diff *manifest.Diff) (interface{}, interface{}) {
	return nil, nil
}
