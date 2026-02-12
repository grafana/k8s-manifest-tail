package manifest

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/telemetry"
	"go.opentelemetry.io/otel/log"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"strings"
)

type Logger struct {
	logger log.Logger
	next   Processor
}

func NewLogger(logger log.Logger, next Processor) *Logger {
	return &Logger{
		logger: logger,
		next:   next,
	}
}

func (l *Logger) Process(rule config.ObjectRule, obj *unstructured.Unstructured, cfg *config.Config) (*Diff, error) {
	var manifestPayload bytes.Buffer
	rawJSON, err := obj.MarshalJSON()
	if err != nil {
		return nil, fmt.Errorf("marshal object: %w", err)
	}
	err = json.Compact(&manifestPayload, rawJSON)
	if err != nil {
		return nil, fmt.Errorf("canonicalize object json: %w", err)
	}

	if obj.GetNamespace() == "" {
		telemetry.Info(
			l.logger,
			manifestPayload.String(),
			log.String("action", "manifest"),
			log.String(fmt.Sprintf("k8s.%s.name", strings.ToLower(obj.GetKind())), obj.GetName()),
		)
	} else {
		telemetry.Info(
			l.logger,
			manifestPayload.String(),
			log.String("action", "manifest"),
			log.String(fmt.Sprintf("k8s.%s.name", strings.ToLower(obj.GetKind())), obj.GetName()),
			log.String("k8s.namespace.name", obj.GetNamespace()),
		)
	}
	return l.next.Process(rule, obj, cfg)
}

func (l *Logger) Delete(rule config.ObjectRule, obj *unstructured.Unstructured, cfg *config.Config) error {
	return l.next.Delete(rule, obj, cfg)
}
