package logging

import (
	"context"
	"testing"

	"github.com/onsi/gomega"

	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/manifest"
	otellog "go.opentelemetry.io/otel/log"
	"go.opentelemetry.io/otel/log/embedded"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNullDiffLoggerDoesNothing(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	logger := NullDiffLogger{}
	logger.Log(nil)
	g.Expect(true).To(gomega.BeTrue()) // no panic/assertion
}

func TestCompactDiffLoggerLogsChanges(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	stub := &stubLogger{}
	logger := NewDiffLogger(config.LoggingConfig{LogDiffs: config.LogDiffsCompact}, stub)
	obj := &unstructured.Unstructured{Object: map[string]interface{}{}}
	obj.SetKind("Pod")
	obj.SetNamespace("default")
	obj.SetName("api")

	logger.Log(&manifest.Diff{Current: obj})

	g.Expect(stub.records).To(gomega.HaveLen(1))
	body := stub.records[0].Body().AsString()
	g.Expect(body).To(gomega.ContainSubstring("Object changed"))
}

type stubLogger struct {
	embedded.Logger
	records []otellog.Record
}

func (s *stubLogger) Emit(_ context.Context, rec otellog.Record) {
	s.records = append(s.records, rec)
}

func (s *stubLogger) Enabled(context.Context, otellog.EnabledParameters) bool {
	return true
}
