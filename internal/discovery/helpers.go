package discovery

import (
	"fmt"
	"slices"

	"github.com/grafana/k8s-manifest-tail/internal/config"
	"k8s.io/apimachinery/pkg/api/meta"
	"k8s.io/apimachinery/pkg/runtime/schema"
)

// ResolveMapping returns the REST mapping for the supplied rule.
func ResolveMapping(mapper meta.RESTMapper, rule config.ObjectRule) (*meta.RESTMapping, error) {
	gv, err := schema.ParseGroupVersion(rule.APIVersion)
	if err != nil {
		return nil, fmt.Errorf("parse apiVersion %q: %w", rule.APIVersion, err)
	}

	gk := schema.GroupKind{
		Group: gv.Group,
		Kind:  rule.Kind,
	}

	mapping, err := mapper.RESTMapping(gk, gv.Version)
	if err != nil {
		return nil, fmt.Errorf("resolve resource for %s (%s): %w", rule.Kind, rule.APIVersion, err)
	}
	return mapping, nil
}

// EffectiveNamespaces returns the namespaces a rule should target.
func EffectiveNamespaces(rule config.ObjectRule, cfg *config.Config) []string {
	if len(rule.Namespaces) > 0 {
		return cloneAndDedupe(rule.Namespaces)
	}
	return cloneAndDedupe(cfg.Namespaces)
}

// ShouldExcludeNamespace reports whether the namespace should be ignored.
func ShouldExcludeNamespace(candidate string, excludedNamespaces []string) bool {
	if excludedNamespaces == nil {
		return false
	}
	for _, namespace := range excludedNamespaces {
		if candidate == namespace {
			return true
		}
	}
	return false
}

func cloneAndDedupe(input []string) []string {
	if len(input) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(input))
	var result []string
	for _, ns := range input {
		if ns == "" {
			continue
		}
		if _, ok := seen[ns]; ok {
			continue
		}
		seen[ns] = struct{}{}
		result = append(result, ns)
	}
	slices.Sort(result)
	return result
}
