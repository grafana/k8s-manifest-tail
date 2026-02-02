package manifest

import (
	"github.com/grafana/k8s-manifest-tail/internal/config"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
)

// Processor handles manifests retrieved from the cluster.
type Processor interface {
	Process(rule config.ObjectRule, obj *unstructured.Unstructured) error
}

func NewProcessor(cfg *config.Config) Processor {
	return NewWriter(cfg.Output)
}
