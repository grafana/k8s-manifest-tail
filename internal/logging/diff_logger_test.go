package logging

import (
	"bytes"
	"testing"

	"github.com/onsi/gomega"

	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/manifest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestNullDiffLoggerDoesNothing(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	logger := NullDiffLogger{}
	logger.Log(&manifest.Diff{})
	g.Expect(true).To(gomega.BeTrue()) // no panic/assertion
}

func TestCompactDiffLoggerLogsChanges(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	var buf bytes.Buffer
	logger := NewDiffLogger(config.LoggingConfig{LogDiffs: config.LogDiffsCompact}, &buf)
	obj := &unstructured.Unstructured{Object: map[string]interface{}{}}
	obj.SetKind("Pod")
	obj.SetNamespace("default")
	obj.SetName("api")

	logger.Log(&manifest.Diff{Current: obj})

	g.Expect(buf.String()).To(gomega.ContainSubstring("Object changed: Pod default/api"))
}
