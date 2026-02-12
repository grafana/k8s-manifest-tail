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
	current := &unstructured.Unstructured{Object: map[string]interface{}{
		"apiVersion": "v1",
		"kind":       "ConfigMap",
		"metadata": map[string]interface{}{
			"name":      "app",
			"namespace": "default",
		},
	}}

	logger.Log(&manifest.Diff{Current: current})

	g.Expect(stub.records).To(gomega.HaveLen(1))
	body := stub.records[0].Body().AsString()

	var payload map[string]interface{}
	g.Expect(json.Unmarshal([]byte(body), &payload)).To(gomega.Succeed())
	g.Expect(payload["kind"]).To(gomega.Equal("ConfigMap"))
	g.Expect(payload["name"]).To(gomega.Equal("app"))
	g.Expect(payload["namespace"]).To(gomega.Equal("default"))
}

func TestGetMinimalDifferenceNilInputs(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	prev, curr := logging.GetMinimalDifference(nil)
	g.Expect(prev).To(gomega.BeNil())
	g.Expect(curr).To(gomega.BeNil())
}

func TestGetMinimalDifferencePrevNil(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	prev, curr := logging.GetMinimalDifference(&manifest.Diff{Current: &unstructured.Unstructured{Object: map[string]interface{}{
		"spec": map[string]interface{}{"replicas": float64(3)},
	}}})
	g.Expect(prev).To(gomega.BeNil())
	g.Expect(curr).To(gomega.HaveKey("spec"))
}

func TestGetMinimalDifferenceCurrNil(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	prev, curr := logging.GetMinimalDifference(&manifest.Diff{Previous: &unstructured.Unstructured{Object: map[string]interface{}{
		"metadata": map[string]interface{}{"labels": map[string]interface{}{"color": "blue"}},
	}}})
	g.Expect(curr).To(gomega.BeNil())
	g.Expect(prev).To(gomega.HaveKey("metadata"))
}

func TestGetMinimalDifferenceIdenticalObjects(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	obj := &unstructured.Unstructured{Object: map[string]interface{}{"spec": map[string]interface{}{"replicas": float64(2)}}}
	prev, curr := logging.GetMinimalDifference(&manifest.Diff{Previous: obj, Current: obj.DeepCopy()})
	g.Expect(prev).To(gomega.BeNil())
	g.Expect(curr).To(gomega.BeNil())
}

func TestGetMinimalDifferenceSimpleFieldChange(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	prev := &unstructured.Unstructured{Object: map[string]interface{}{"spec": map[string]interface{}{"replicas": float64(2)}}}
	curr := &unstructured.Unstructured{Object: map[string]interface{}{"spec": map[string]interface{}{"replicas": float64(3)}}}
	prevDiff, currDiff := logging.GetMinimalDifference(&manifest.Diff{Previous: prev, Current: curr})
	g.Expect(prevDiff).To(gomega.Equal(map[string]interface{}{"spec": map[string]interface{}{"replicas": float64(2)}}))
	g.Expect(currDiff).To(gomega.Equal(map[string]interface{}{"spec": map[string]interface{}{"replicas": float64(3)}}))
}

func TestGetMinimalDifferenceNestedChange(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	prev := &unstructured.Unstructured{Object: map[string]interface{}{"metadata": map[string]interface{}{"labels": map[string]interface{}{"color": "blue", "team": "alpha"}}}}
	curr := &unstructured.Unstructured{Object: map[string]interface{}{"metadata": map[string]interface{}{"labels": map[string]interface{}{"color": "green", "team": "alpha"}}}}
	prevDiff, currDiff := logging.GetMinimalDifference(&manifest.Diff{Previous: prev, Current: curr})
	g.Expect(prevDiff).To(gomega.Equal(map[string]interface{}{"metadata": map[string]interface{}{"labels": map[string]interface{}{"color": "blue"}}}))
	g.Expect(currDiff).To(gomega.Equal(map[string]interface{}{"metadata": map[string]interface{}{"labels": map[string]interface{}{"color": "green"}}}))
}

func TestGetMinimalDifferenceRemovedField(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	prev := &unstructured.Unstructured{Object: map[string]interface{}{"metadata": map[string]interface{}{"labels": map[string]interface{}{"color": "blue", "env": "prod"}}}}
	curr := &unstructured.Unstructured{Object: map[string]interface{}{"metadata": map[string]interface{}{"labels": map[string]interface{}{"color": "blue"}}}}
	prevDiff, currDiff := logging.GetMinimalDifference(&manifest.Diff{Previous: prev, Current: curr})
	g.Expect(prevDiff).To(gomega.Equal(map[string]interface{}{"metadata": map[string]interface{}{"labels": map[string]interface{}{"env": "prod"}}}))
	g.Expect(currDiff).To(gomega.BeNil())
}

func TestGetMinimalDifferenceListChange(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	prev := &unstructured.Unstructured{Object: map[string]interface{}{"spec": map[string]interface{}{"containers": []interface{}{"a"}}}}
	curr := &unstructured.Unstructured{Object: map[string]interface{}{"spec": map[string]interface{}{"containers": []interface{}{"a", "b"}}}}
	prevDiff, currDiff := logging.GetMinimalDifference(&manifest.Diff{Previous: prev, Current: curr})
	g.Expect(prevDiff).To(gomega.Equal(map[string]interface{}{"spec": map[string]interface{}{"containers": []interface{}{"a"}}}))
	g.Expect(currDiff).To(gomega.Equal(map[string]interface{}{"spec": map[string]interface{}{"containers": []interface{}{"a", "b"}}}))
}

func TestGetMinimalDifferenceComplexChange(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)
	prev := &unstructured.Unstructured{Object: map[string]interface{}{
		"metadata": map[string]interface{}{"labels": map[string]interface{}{"color": "blue"}},
		"spec":     map[string]interface{}{"replicas": float64(2)},
	}}
	curr := &unstructured.Unstructured{Object: map[string]interface{}{
		"metadata": map[string]interface{}{"labels": map[string]interface{}{"color": "green", "team": "platform"}},
		"spec":     map[string]interface{}{"replicas": float64(3)},
	}}
	prevDiff, currDiff := logging.GetMinimalDifference(&manifest.Diff{Previous: prev, Current: curr})
	g.Expect(prevDiff).To(gomega.Equal(map[string]interface{}{
		"metadata": map[string]interface{}{"labels": map[string]interface{}{"color": "blue"}},
		"spec":     map[string]interface{}{"replicas": float64(2)},
	}))
	g.Expect(currDiff).To(gomega.Equal(map[string]interface{}{
		"metadata": map[string]interface{}{"labels": map[string]interface{}{"color": "green", "team": "platform"}},
		"spec":     map[string]interface{}{"replicas": float64(3)},
	}))
}
