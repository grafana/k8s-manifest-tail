package logging_test

import (
	"encoding/json"
	"testing"

	"github.com/onsi/gomega"

	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/logging"
	"github.com/grafana/k8s-manifest-tail/internal/manifest"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

func TestDetailedDiffLoggerEmitsJSON(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	stub := &StubLogger{}
	logger := logging.NewDiffLogger(config.LoggingConfig{LogDiffs: config.LogDiffsDetailed}, stub)
	current := &unstructured.Unstructured{}
	current.SetAPIVersion("v1")
	current.SetKind("ConfigMap")
	current.SetNamespace("default")
	current.SetName("app")

	logger.Log(&manifest.Diff{
		Previous: nil,
		Current:  current,
	})

	g.Expect(stub.records).To(gomega.HaveLen(1))
	body := stub.records[0].Body().AsString()

	var payload logging.DetailedDiffReport
	g.Expect(json.Unmarshal([]byte(body), &payload)).To(gomega.Succeed())
	g.Expect(payload.Kind).To(gomega.Equal("ConfigMap"))
	g.Expect(payload.Name).To(gomega.Equal("app"))
	g.Expect(payload.Namespace).To(gomega.Equal("default"))
	g.Expect(payload.Action).To(gomega.Equal("created"))
}
