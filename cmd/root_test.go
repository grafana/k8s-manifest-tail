package cmd

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega"
)

func TestLoadConfigWithOverrides(t *testing.T) {
	state := snapshotFlags()
	defer restoreFlags(state)

	g := gomega.NewWithT(t)

	configPath = writeTempConfigFile(t, `
output:
  directory: from-config
  format: yaml
refreshInterval: 6h
namespaces: ["default"]
objects:
  - apiVersion: v1
    kind: Pod
`)

	kubeconfigOverride = "/tmp/flag-kubeconfig"
	outputDirOverride = "custom-dir"
	outputFormatOverride = "json"
	refreshIntervalOverride = "2h"
	namespacesOverride = []string{"prod"}
	excludeNamespacesOverride = []string{"kube-system"}

	err := LoadConfiguration(nil, nil)

	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(Configuration.KubeconfigPath).To(gomega.Equal("/tmp/flag-kubeconfig"))
	g.Expect(Configuration.Output.Directory).To(gomega.Equal("custom-dir"))
	g.Expect(string(Configuration.Output.Format)).To(gomega.Equal("json"))
	g.Expect(Configuration.RefreshInterval).To(gomega.Equal("2h"))
	g.Expect(Configuration.Namespaces).To(gomega.Equal([]string{"prod"}))
	g.Expect(Configuration.ExcludeNamespaces).To(gomega.Equal([]string{"kube-system"}))
}

func TestLoadConfigWithOverridesInvalidOutputFormat(t *testing.T) {
	state := snapshotFlags()
	defer restoreFlags(state)

	g := gomega.NewWithT(t)

	configPath = writeTempConfigFile(t, `
output:
  directory: out
  format: foo
objects:
  - apiVersion: v1
    kind: Service
`)

	err := LoadConfiguration(nil, nil)

	g.Expect(err).To(gomega.HaveOccurred())
	g.Expect(err.Error()).To(gomega.Equal("invalid output format \"foo\" (expected yaml or json)"))
}

type flagState struct {
	configPath                string
	kubeconfigOverride        string
	outputDirOverride         string
	outputFormatOverride      string
	refreshIntervalOverride   string
	namespacesOverride        []string
	excludeNamespacesOverride []string
}

func snapshotFlags() flagState {
	return flagState{
		configPath:                configPath,
		kubeconfigOverride:        kubeconfigOverride,
		outputDirOverride:         outputDirOverride,
		outputFormatOverride:      outputFormatOverride,
		refreshIntervalOverride:   refreshIntervalOverride,
		namespacesOverride:        cloneSlice(namespacesOverride),
		excludeNamespacesOverride: cloneSlice(excludeNamespacesOverride),
	}
}

func restoreFlags(state flagState) {
	configPath = state.configPath
	kubeconfigOverride = state.kubeconfigOverride
	outputDirOverride = state.outputDirOverride
	outputFormatOverride = state.outputFormatOverride
	refreshIntervalOverride = state.refreshIntervalOverride
	namespacesOverride = cloneSlice(state.namespacesOverride)
	excludeNamespacesOverride = cloneSlice(state.excludeNamespacesOverride)
}

func cloneSlice(values []string) []string {
	if values == nil {
		return nil
	}
	out := make([]string, len(values))
	copy(out, values)
	return out
}

func writeTempConfigFile(t *testing.T, contents string) string {
	t.Helper()

	dir := t.TempDir()
	path := filepath.Join(dir, "config.yaml")
	if err := os.WriteFile(path, []byte(contents), 0o600); err != nil {
		t.Fatalf("write temp config: %v", err)
	}
	return path
}
