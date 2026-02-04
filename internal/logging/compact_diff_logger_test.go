package logging_test

import (
	"github.com/grafana/k8s-manifest-tail/internal/logging"
	"testing"

	"github.com/onsi/gomega"

	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/manifest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestCompactDiffLoggerLogsChanges(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	stub := &StubLogger{}
	logger := logging.NewDiffLogger(config.LoggingConfig{LogDiffs: config.LogDiffsCompact}, stub)
	obj := &unstructured.Unstructured{Object: map[string]interface{}{}}
	obj.SetKind("Pod")
	obj.SetNamespace("default")
	obj.SetName("api")

	logger.Log(&manifest.Diff{Current: obj})

	g.Expect(stub.records).To(gomega.HaveLen(1))
	body := stub.records[0].Body().AsString()
	g.Expect(body).To(gomega.ContainSubstring("Object created: Pod default/api"))
}
