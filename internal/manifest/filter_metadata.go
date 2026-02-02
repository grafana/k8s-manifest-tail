package manifest

import "k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

// RemoveMetadataFieldsFilter strips server-populated metadata entries.
type RemoveMetadataFieldsFilter struct{}

// Apply removes known noisy metadata fields.
func (RemoveMetadataFieldsFilter) Apply(obj *unstructured.Unstructured) error {
	if obj == nil {
		return nil
	}

	fields := []string{"managedFields", "resourceVersion", "uid", "selfLink", "generation", "creationTimestamp"}
	for _, field := range fields {
		unstructured.RemoveNestedField(obj.Object, "metadata", field)
	}
	return nil
}
