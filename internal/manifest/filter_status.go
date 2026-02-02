package manifest

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// RemoveStatusFilter strips the status stanza from manifests.
type RemoveStatusFilter struct{}

// Apply removes the status field from the given object.
func (RemoveStatusFilter) Apply(obj *unstructured.Unstructured) error {
	if obj == nil {
		return nil
	}
	unstructured.RemoveNestedField(obj.Object, "status")
	return nil
}
