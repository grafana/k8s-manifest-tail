package manifest

import (
	"bytes"
	"context"
	"encoding/json"
	"testing"

	"github.com/onsi/gomega"
	"go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/embedded"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/grafana/k8s-manifest-tail/internal/config"
)

func TestLoggerEmitsManifestWithNamespace(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	stubLogger := &capturingLogger{}
	next := &stubManifestProcessor{}
	logger := NewLogger(stubLogger, next)

	obj := newUnstructured("v1", "ConfigMap", "default", "settings")
	_, err := logger.Process(config.ObjectRule{Kind: "ConfigMap"}, obj, &config.Config{})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(next.processed).To(gomega.HaveLen(1))
	g.Expect(stubLogger.records).To(gomega.HaveLen(1))

	record := stubLogger.records[0]
	g.Expect(record.Severity()).To(gomega.Equal(log.SeverityInfo))
	rawJSON, marshalErr := obj.MarshalJSON()
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
	next := &stubManifestProcessor{}
	logger := NewLogger(stubLogger, next)

	obj := newUnstructured("v1", "Node", "", "node-a")
	_, err := logger.Process(config.ObjectRule{Kind: "Node"}, obj, &config.Config{})
	g.Expect(err).NotTo(gomega.HaveOccurred())

	g.Expect(stubLogger.records).To(gomega.HaveLen(1))
	attrs := recordAttributes(stubLogger.records[0])
	g.Expect(attrs).To(gomega.HaveKeyWithValue("k8s.node.name", "node-a"))
	g.Expect(attrs).NotTo(gomega.HaveKey("k8s.namespace.name"))

	g.Expect(next.processed).To(gomega.HaveLen(1))

	err = logger.Delete(config.ObjectRule{Kind: "Node"}, obj, &config.Config{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(next.deleted).To(gomega.HaveLen(1))
	g.Expect(next.deleted[0].GetName()).To(gomega.Equal("node-a"))
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

type stubManifestProcessor struct {
	processed []*unstructured.Unstructured
	deleted   []*unstructured.Unstructured
}

func (s *stubManifestProcessor) Process(_ config.ObjectRule, obj *unstructured.Unstructured, _ *config.Config) (*Diff, error) {
	s.processed = append(s.processed, obj.DeepCopy())
	return &Diff{Current: obj.DeepCopy()}, nil
}

func (s *stubManifestProcessor) Delete(_ config.ObjectRule, obj *unstructured.Unstructured, _ *config.Config) error {
	s.deleted = append(s.deleted, obj.DeepCopy())
	return nil
}

func recordAttributes(rec log.Record) map[string]string {
	attrs := make(map[string]string, rec.AttributesLen())
	rec.WalkAttributes(func(kv log.KeyValue) bool {
		attrs[kv.Key] = kv.Value.AsString()
		return true
	})
	return attrs
}
