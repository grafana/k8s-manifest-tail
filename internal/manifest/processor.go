package manifest

import (
	"github.com/grafana/k8s-manifest-tail/internal/config"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Diff captures the before/after state for a manifest write.
type Diff struct {
	Previous *unstructured.Unstructured
	Current  *unstructured.Unstructured
}

// Processor handles manifests retrieved from the cluster.
type Processor interface {
	Process(rule config.ObjectRule, obj *unstructured.Unstructured, cfg *config.Config) (*Diff, error)
	Delete(rule config.ObjectRule, obj *unstructured.Unstructured, cfg *config.Config) error
}
