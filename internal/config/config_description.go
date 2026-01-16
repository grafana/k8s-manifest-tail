package config

import (
	"fmt"
	"strings"
)

func (c *Config) Describe() string {
	result := fmt.Sprintf("This configuration will get manifests for:\n")
	for _, rule := range c.Objects {
		result += fmt.Sprintf("  %s\n", rule.Describe(c))
	}
	return result
}

func (rule *ObjectRule) Describe(c *Config) string {
	pluralKind := pluralizeKind(rule.Kind)
	includedNamespaces := c.Namespaces
	if len(rule.Namespaces) > 0 {
		includedNamespaces = rule.Namespaces
	}
	excludedNamespaces := c.ExcludeNamespaces
	return fmt.Sprintf("%s in %s", pluralKind, namespaceList(includedNamespaces, excludedNamespaces))
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

func namespaceList(included, excluded []string) string {
	result := ""
	if len(included) == 0 {
		result = "all namespaces"
	} else {
		result = formatList(included, "or")
	}
	if len(excluded) > 0 {
		result = fmt.Sprintf("%s excluding %s", result, formatList(included, "and"))
	}
	return result
}

func formatList(namespaces []string, conjunction string) string {
	quoted := make([]string, len(namespaces))
	for i, ns := range namespaces {
		quoted[i] = fmt.Sprintf("%q", ns)
	}
	if len(quoted) == 1 {
		return fmt.Sprintf("the %s namespace", quoted[0])
	} else if len(quoted) == 2 {
		return fmt.Sprintf("the %s %s %s namespaces", quoted[0], conjunction, quoted[1])
	}
	return fmt.Sprintf("the %s %s %s namespaces", strings.Join(quoted[:len(quoted)-1], ", "), conjunction, quoted[len(quoted)-1])
}
