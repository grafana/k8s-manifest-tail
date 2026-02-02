package manifest

import (
	"testing"

	"github.com/onsi/gomega"
)

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
	g.Expect(obj.Object["metadata"]).NotTo(gomega.HaveKey("resourceVersion"))
	g.Expect(obj.Object["metadata"]).NotTo(gomega.HaveKey("uid"))
	g.Expect(obj.Object["metadata"]).NotTo(gomega.HaveKey("selfLink"))
	g.Expect(obj.Object["metadata"]).NotTo(gomega.HaveKey("generation"))
	g.Expect(obj.Object["metadata"]).NotTo(gomega.HaveKey("creationTimestamp"))
}
