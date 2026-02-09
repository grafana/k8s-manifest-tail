package discovery

import (
	"context"
	"fmt"
	"regexp"
	"strings"

	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"

	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/kube"
)

// Fetcher lists Kubernetes resources according to configuration rules.
type Fetcher struct {
	clients *kube.Clients
	cfg     *config.Config
}

// NewFetcher builds a Fetcher that uses the supplied clients and configuration.
func NewFetcher(clients *kube.Clients, cfg *config.Config) *Fetcher {
	return &Fetcher{
		clients: clients,
		cfg:     cfg,
	}
}

// FetchResources returns all objects that match a single rule.
func (f *Fetcher) FetchResources(ctx context.Context, rule config.ObjectRule) ([]unstructured.Unstructured, error) {
	mapping, err := MappingFromRule(f.clients.Mapper, rule)
	if err != nil {
		return nil, err
	}

	resourceClient := f.clients.Dynamic.Resource(mapping.Resource)

	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		return f.fetchNamespaced(ctx, resourceClient, rule)
	}
	return f.fetchClusterScoped(ctx, resourceClient, rule)
}

// fetchClusterScoped returns all Cluster-scoped objects that match a single rule.
func (f *Fetcher) fetchClusterScoped(ctx context.Context, client dynamic.NamespaceableResourceInterface, rule config.ObjectRule) ([]unstructured.Unstructured, error) {
	list, err := client.Namespace("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, fmt.Errorf("list %s (cluster-scoped): %w", rule.Kind, err)
	}
	filtered, err := filterByNamePattern(list.Items, rule.NamePattern)
	if err != nil {
		return nil, fmt.Errorf("filter %s by name: %w", rule.Kind, err)
	}
	return filtered, nil
}

// fetchNamespaced returns all namespaced objects that match a single rule.
func (f *Fetcher) fetchNamespaced(ctx context.Context, client dynamic.NamespaceableResourceInterface, rule config.ObjectRule) ([]unstructured.Unstructured, error) {
	namespaces := f.cfg.Namespaces
	if len(rule.Namespaces) > 0 {
		namespaces = rule.Namespaces
	}

	if len(namespaces) == 0 {
		list, err := client.Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("list %s across namespaces: %w", rule.Kind, err)
		}
		filtered := filterExcluded(list.Items, f.cfg.ExcludeNamespaces)
		return applyNameFilter(filtered, rule, rule.Kind)
	}

	var all []unstructured.Unstructured
	for _, ns := range namespaces {
		list, err := client.Namespace(ns).List(ctx, metav1.ListOptions{})
		if err != nil {
			return nil, fmt.Errorf("list %s in namespace %s: %w", rule.Kind, ns, err)
		}
		all = append(all, list.Items...)
	}
	return applyNameFilter(all, rule, rule.Kind)
}

func MappingFromRule(mapper meta.RESTMapper, rule config.ObjectRule) (*meta.RESTMapping, error) {
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

func filterExcluded(items []unstructured.Unstructured, excludedNamespaces []string) []unstructured.Unstructured {
	if len(excludedNamespaces) == 0 {
		return items
	}
	var filtered []unstructured.Unstructured
	for _, item := range items {
		isExcluded := false
		for _, excludedNamespace := range excludedNamespaces {
			if item.GetNamespace() == excludedNamespace {
				isExcluded = true
				break
			}
		}
		if !isExcluded {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func applyNameFilter(items []unstructured.Unstructured, rule config.ObjectRule, kind string) ([]unstructured.Unstructured, error) {
	filtered, err := filterByNamePattern(items, rule.NamePattern)
	if err != nil {
		return nil, fmt.Errorf("filter %s by name: %w", kind, err)
	}
	return filtered, nil
}

func filterByNamePattern(items []unstructured.Unstructured, pattern string) ([]unstructured.Unstructured, error) {
	pattern = strings.TrimSpace(pattern)
	if pattern == "" {
		return items, nil
	}
	re, err := regexp.Compile(pattern)
	if err != nil {
		return nil, fmt.Errorf("compile name pattern: %w", err)
	}
	var filtered []unstructured.Unstructured
	for _, item := range items {
		if re.MatchString(item.GetName()) {
			filtered = append(filtered, item)
		}
	}
	return filtered, nil
}
