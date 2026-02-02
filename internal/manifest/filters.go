package manifest

import (
	"fmt"

	"github.com/grafana/k8s-manifest-tail/internal/config"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Filter mutates a manifest prior to persistence.
type Filter interface {
	Apply(obj *unstructured.Unstructured) error
}

// FilterProcessor runs filters before delegating to the next processor.
type FilterProcessor struct {
	next    Processor
	filters []Filter
}

// NewFilterProcessor constructs a processor that applies filters before invoking next.
func NewFilterProcessor(next Processor, filters ...Filter) Processor {
	return &FilterProcessor{
		next:    next,
		filters: filters,
	}
}

// Process applies filters and passes the object to the next processor.
func (p *FilterProcessor) Process(rule config.ObjectRule, obj *unstructured.Unstructured, cfg *config.Config) (*Diff, error) {
	for _, filter := range p.filters {
		if err := filter.Apply(obj); err != nil {
			return nil, fmt.Errorf("apply filter: %w", err)
		}
	}
	return p.next.Process(rule, obj, cfg)
}

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

// RemoveMetadataFieldsFilter strips server-populated metadata entries.
type RemoveMetadataFieldsFilter struct{}

// Apply removes known noisy metadata fields.
func (RemoveMetadataFieldsFilter) Apply(obj *unstructured.Unstructured) error {
	if obj == nil {
		return nil
	}
	fields := []string{"managedFields", "uid", "selfLink"}
	for _, field := range fields {
		unstructured.RemoveNestedField(obj.Object, "metadata", field)
	}
	return nil
}
