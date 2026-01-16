package config

import (
	"testing"

	"github.com/onsi/gomega"
)

func TestDescribe_OneObject_AllNamespaces(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	cfg := &Config{
		Namespaces:        []string{},
		ExcludeNamespaces: []string{},
		Objects: []ObjectRule{
			{
				APIVersion: "v1",
				Kind:       "Pod",
			},
		},
	}

	description := cfg.Describe()
	g.Expect(description).To(gomega.ContainSubstring("This configuration will get manifests for:"))
	g.Expect(description).To(gomega.ContainSubstring("  Pods in all namespaces"))
}

func TestDescribe_MultipleObjectsWithGlobalNamespaces(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	cfg := &Config{
		Namespaces: []string{"default", "prod", "staging"},
		Objects: []ObjectRule{
			{APIVersion: "v1", Kind: "Pod"},
			{APIVersion: "apps/v1", Kind: "Deployment"},
		},
	}

	description := cfg.Describe()
	g.Expect(description).To(gomega.ContainSubstring("This configuration will get manifests for:"))
	g.Expect(description).To(gomega.ContainSubstring(`  Pods in the "default", "prod", or "staging" namespaces`))
	g.Expect(description).To(gomega.ContainSubstring(`  Deployments in the "default", "prod", or "staging" namespaces`))
}

func TestDescribe_ObjectLevelNamespacesOverrideGlobal(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	cfg := &Config{
		Namespaces: []string{"default"},
		Objects: []ObjectRule{
			{APIVersion: "apps/v1", Kind: "Deployment"},
			{APIVersion: "v1", Kind: "Service", Namespaces: []string{"prod", "staging"}},
		},
	}

	description := cfg.Describe()
	g.Expect(description).To(gomega.ContainSubstring("This configuration will get manifests for:"))
	g.Expect(description).To(gomega.ContainSubstring(`  Deployments in the "default" namespace`))
	g.Expect(description).To(gomega.ContainSubstring(`  Services in the "prod" or "staging" namespaces`))
}

func TestDescribe_AllNamespacesWithExclusions(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	cfg := &Config{
		ExcludeNamespaces: []string{"kube-system", "observability"},
		Objects: []ObjectRule{
			{APIVersion: "v1", Kind: "Pod"},
		},
	}

	description := cfg.Describe()
	g.Expect(description).To(gomega.ContainSubstring("This configuration will get manifests for:"))
	g.Expect(description).To(gomega.ContainSubstring(`  Pods in all namespaces except "kube-system" or "observability"`))
}

func TestDescribe_IncludedNamespacesWithExclusions(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	cfg := &Config{
		Namespaces:        []string{"default", "prod"},
		ExcludeNamespaces: []string{"kube-system"},
		Objects: []ObjectRule{
			{APIVersion: "apps/v1", Kind: "Deployment"},
		},
	}

	description := cfg.Describe()
	g.Expect(description).To(gomega.ContainSubstring("This configuration will get manifests for:"))
	g.Expect(description).To(gomega.ContainSubstring(`  Deployments in the "default" or "prod" namespaces (excluding "kube-system")`))
}
