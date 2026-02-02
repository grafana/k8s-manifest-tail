package manifest

import (
	"testing"

	"github.com/onsi/gomega"

	"github.com/grafana/k8s-manifest-tail/internal/config"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

type stubProcessor struct {
	lastObj *unstructured.Unstructured
}

func (s *stubProcessor) Process(rule config.ObjectRule, obj *unstructured.Unstructured, cfg *config.Config) (*Diff, error) {
	s.lastObj = obj
	return nil, nil
}

func TestFilterProcessorRemovesStatus(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	next := &stubProcessor{}
	processor := NewFilterProcessor(next, RemoveStatusFilter{})

	obj := newUnstructured("v1", "Pod", "default", "api")
	obj.Object["status"] = map[string]interface{}{"phase": "Running"}

	_, err := processor.Process(config.ObjectRule{Kind: "Pod"}, obj, &config.Config{})
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(next.lastObj.Object).NotTo(gomega.HaveKey("status"))
}

func TestRemoveStatusFilterHandlesNil(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	filter := RemoveStatusFilter{}
	err := filter.Apply(nil)
	g.Expect(err).NotTo(gomega.HaveOccurred())
}

func TestRemoveMetadataFieldsFilterStripsKnownFields(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	obj := newUnstructured("v1", "Service", "default", "api")
	metadata := obj.Object["metadata"].(map[string]interface{})
	metadata["managedFields"] = []interface{}{"something"}
	metadata["resourceVersion"] = "123"
	metadata["uid"] = "uid-123"
	metadata["selfLink"] = "/api/v1"
	metadata["generation"] = int64(2)
	metadata["creationTimestamp"] = "2024-01-01T00:00:00Z"

	filter := RemoveMetadataFieldsFilter{}
	err := filter.Apply(obj)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(obj.Object["metadata"]).NotTo(gomega.HaveKey("managedFields"))
	g.Expect(obj.Object["metadata"]).NotTo(gomega.HaveKey("uid"))
	g.Expect(obj.Object["metadata"]).NotTo(gomega.HaveKey("selfLink"))
}
