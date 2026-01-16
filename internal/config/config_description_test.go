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

	g.Expect(cfg.Describe()).To(gomega.ContainSubstring("This configuration will get manifests for:"))
	g.Expect(cfg.Describe()).To(gomega.ContainSubstring("  Pods in all namespaces"))
}
