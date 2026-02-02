package manifest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega"

	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/grafana/k8s-manifest-tail/internal/config"
)

func TestWriterWritesYAMLManifest(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	dir := t.TempDir()
	writer := NewWriter(config.OutputConfig{
		Directory: dir,
		Format:    config.OutputFormatYAML,
	})

	obj := newUnstructured("v1", "Pod", "default", "api")
	err := writer.Process(config.ObjectRule{Kind: "Pod"}, obj)

	g.Expect(err).NotTo(gomega.HaveOccurred())
	path := filepath.Join(dir, "Pod", "default", "api.yaml")
	g.Expect(path).To(gomega.BeAnExistingFile())

	contents, readErr := os.ReadFile(path)
	g.Expect(readErr).NotTo(gomega.HaveOccurred())
	g.Expect(string(contents)).To(gomega.ContainSubstring("kind: Pod"))
	g.Expect(string(contents)).To(gomega.HaveSuffix("\n"))
}

func TestWriterWritesJSONManifest(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	dir := t.TempDir()
	writer := NewWriter(config.OutputConfig{
		Directory: dir,
		Format:    config.OutputFormatJSON,
	})

	obj := newUnstructured("apps/v1", "Deployment", "prod", "frontend")
	err := writer.Process(config.ObjectRule{Kind: "Deployment"}, obj)

	g.Expect(err).NotTo(gomega.HaveOccurred())
	path := filepath.Join(dir, "Deployment", "prod", "frontend.json")
	g.Expect(path).To(gomega.BeAnExistingFile())

	contents, readErr := os.ReadFile(path)
	g.Expect(readErr).NotTo(gomega.HaveOccurred())
	g.Expect(string(contents)).To(gomega.ContainSubstring(`"kind": "Deployment"`))
	g.Expect(string(contents)).To(gomega.HaveSuffix("\n"))
}

func TestWriterUsesClusterDirectoryForClusterScopedResources(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	dir := t.TempDir()
	writer := NewWriter(config.OutputConfig{
		Directory: dir,
		Format:    config.OutputFormatYAML,
	})

	obj := newUnstructured("v1", "Node", "", "node-a")
	err := writer.Process(config.ObjectRule{Kind: "Node"}, obj)

	g.Expect(err).NotTo(gomega.HaveOccurred())
	path := filepath.Join(dir, "Node", "cluster", "node-a.yaml")
	g.Expect(path).To(gomega.BeAnExistingFile())
}

func newUnstructured(apiVersion, kind, namespace, name string) *unstructured.Unstructured {
	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion(apiVersion)
	obj.SetKind(kind)
	obj.SetNamespace(namespace)
	obj.SetName(name)
	obj.Object = map[string]interface{}{
		"apiVersion": apiVersion,
		"kind":       kind,
		"metadata": map[string]interface{}{
			"name":      name,
			"namespace": namespace,
		},
	}
	return obj
}
