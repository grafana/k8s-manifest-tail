package logging

import (
	"bytes"
	"context"
	"encoding/json"
	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/manifest"
	"testing"

	"github.com/onsi/gomega"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/embedded"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestLoggerEmitsManifestWithNamespace(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	stubLogger := &capturingLogger{}
	logger := NewManifestLogger(config.LoggingConfig{LogManifests: true}, stubLogger)

	current := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name":      "settings",
			"namespace": "default",
		},
	}}

	logger.Log(&manifest.Diff{Current: current})

	g.Expect(stubLogger.records).To(gomega.HaveLen(1))

	record := stubLogger.records[0]
	g.Expect(record.Severity()).To(gomega.Equal(log.SeverityInfo))
	rawJSON, marshalErr := current.MarshalJSON()
	g.Expect(marshalErr).NotTo(gomega.HaveOccurred())
	var compact bytes.Buffer
	g.Expect(json.Compact(&compact, rawJSON)).To(gomega.Succeed())
	g.Expect(record.Body().AsString()).To(gomega.Equal(compact.String()))

	attrs := recordAttributes(record)
	g.Expect(attrs).To(gomega.HaveKeyWithValue("k8s.configmap.name", "settings"))
	g.Expect(attrs).To(gomega.HaveKeyWithValue("k8s.namespace.name", "default"))
}

func TestLoggerOmitsNamespaceForClusterScopedObjects(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	stubLogger := &capturingLogger{}
	logger := NewManifestLogger(config.LoggingConfig{LogManifests: true}, stubLogger)

	current := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "Node",
		"metadata": map[string]interface{}{
			"name": "node-a",
		},
	}}

	logger.Log(&manifest.Diff{Current: current})

	g.Expect(stubLogger.records).To(gomega.HaveLen(1))
	attrs := recordAttributes(stubLogger.records[0])
	g.Expect(attrs).To(gomega.HaveKeyWithValue("k8s.node.name", "node-a"))
	g.Expect(attrs).NotTo(gomega.HaveKey("k8s.namespace.name"))
}

type capturingLogger struct {
	embedded.Logger
	records []log.Record
}

func (c *capturingLogger) Emit(_ context.Context, record log.Record) {
	c.records = append(c.records, record.Clone())
}

func (c *capturingLogger) Enabled(context.Context, log.EnabledParameters) bool {
	return true
}

func recordAttributes(rec log.Record) map[string]string {
	attrs := make(map[string]string, rec.AttributesLen())
	rec.WalkAttributes(func(kv log.KeyValue) bool {
		attrs[kv.Key] = kv.Value.AsString()
		return true
	})
	return attrs
}
