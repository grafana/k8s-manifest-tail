package config

import (
	"fmt"
	"github.com/grafana/k8s-manifest-tail/internal"
	"strings"
)

func (cfg *Config) Describe() string {
	result := "This configuration will get manifests for:\n"
	for _, rule := range cfg.Objects {
		result += fmt.Sprintf("  %s\n", rule.Describe(cfg))
	}
	return result
}

func (rule *ObjectRule) Describe(c *Config) string {
	pluralKind := pluralizeKind(rule.Kind)
	includedNamespaces := rule.Namespaces
	if len(includedNamespaces) == 0 {
		includedNamespaces = c.Namespaces
	}
	return fmt.Sprintf("%s in %s", pluralKind, describeNamespaceScope(includedNamespaces, c.ExcludeNamespaces))
}

func pluralizeKind(kind string) string {
	if kind == "" {
		return "Objects"
	}
	if strings.HasSuffix(strings.ToLower(kind), "s") {
		return kind
	}
	return kind + "s"
}

func describeNamespaceScope(included, excluded []string) string {
	if len(included) == 0 {
		if len(excluded) == 0 {
			return "all namespaces"
		}
		return fmt.Sprintf("all namespaces except %s", internal.FormatQuotedList(excluded))
	}

	scope := "namespaces"
	if len(included) == 1 {
		scope = "namespace"
	}

	description := fmt.Sprintf("the %s %s", internal.FormatQuotedList(included), scope)
	if len(excluded) > 0 {
		description = fmt.Sprintf("%s (excluding %s)", description, internal.FormatQuotedList(excluded))
	}
	return description
}
