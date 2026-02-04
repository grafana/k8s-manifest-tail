package logging_test

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/logging"
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

func TestCompactDiffLoggerLogsDeletion(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	stub := &StubLogger{}
	logger := logging.NewDiffLogger(config.LoggingConfig{LogDiffs: config.LogDiffsCompact}, stub)
	prev := &unstructured.Unstructured{}
	prev.SetKind("Pod")
	prev.SetNamespace("default")
	prev.SetName("api")

	logger.Log(&manifest.Diff{Previous: prev})

	g.Expect(stub.records).To(gomega.HaveLen(1))
	body := stub.records[0].Body().AsString()
	g.Expect(body).To(gomega.ContainSubstring("Object deleted: Pod default/api"))
}

func TestCompactDiffLoggerLogsModification(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	stub := &StubLogger{}
	logger := logging.NewDiffLogger(config.LoggingConfig{LogDiffs: config.LogDiffsCompact}, stub)
	prev := &unstructured.Unstructured{}
	prev.SetKind("Pod")
	prev.SetNamespace("default")
	prev.SetName("api")
	current := &unstructured.Unstructured{}
	current.SetKind("Pod")
	current.SetNamespace("default")
	current.SetName("api")

	logger.Log(&manifest.Diff{Previous: prev, Current: current})

	g.Expect(stub.records).To(gomega.HaveLen(1))
	body := stub.records[0].Body().AsString()
	g.Expect(body).To(gomega.ContainSubstring("Object modified: Pod default/api"))
}

func TestCompactDiffLoggerLogsClusterScopedCreation(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	stub := &StubLogger{}
	logger := logging.NewDiffLogger(config.LoggingConfig{LogDiffs: config.LogDiffsCompact}, stub)
	clusterRole := &unstructured.Unstructured{}
	clusterRole.SetKind("ClusterRole")
	clusterRole.SetName("my-cluster-role")

	logger.Log(&manifest.Diff{Current: clusterRole})

	g.Expect(stub.records).To(gomega.HaveLen(1))
	body := stub.records[0].Body().AsString()
	g.Expect(body).To(gomega.ContainSubstring("Object created: ClusterRole my-cluster-role"))
}
