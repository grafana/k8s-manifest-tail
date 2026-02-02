package manifest

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/onsi/gomega"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"

	"github.com/grafana/k8s-manifest-tail/internal/config"
)

func TestWriterWritesYAMLManifestAndReportsDiff(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	dir := t.TempDir()
	writer := NewWriter(config.OutputConfig{
		Directory: dir,
		Format:    config.OutputFormatYAML,
	})

	obj := newUnstructured("v1", "Pod", "default", "api")
	diff, err := writer.Process(config.ObjectRule{Kind: "Pod"}, obj, nil)

	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(diff).NotTo(gomega.BeNil())
	g.Expect(diff.Previous).To(gomega.BeNil())
	g.Expect(diff.Current.GetName()).To(gomega.Equal("api"))

	path := filepath.Join(dir, "Pod", "default", "api.yaml")
	g.Expect(path).To(gomega.BeAnExistingFile())

	content, readErr := os.ReadFile(path)
	g.Expect(readErr).NotTo(gomega.HaveOccurred())
	g.Expect(string(content)).To(gomega.ContainSubstring("kind: Pod"))
	g.Expect(string(content)).To(gomega.HaveSuffix("\n"))
}

func TestWriterWritesJSONManifestAndDetectsUpdates(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	dir := t.TempDir()
	writer := NewWriter(config.OutputConfig{
		Directory: dir,
		Format:    config.OutputFormatJSON,
	})

	rule := config.ObjectRule{Kind: "Deployment"}
	first := newUnstructured("apps/v1", "Deployment", "prod", "frontend")
	diff, err := writer.Process(rule, first, nil)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(diff).NotTo(gomega.BeNil())

	second := newUnstructured("apps/v1", "Deployment", "prod", "frontend")
	second.Object["metadata"].(map[string]interface{})["annotations"] = map[string]interface{}{"team": "edge"}
	diff, err = writer.Process(rule, second, nil)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(diff).NotTo(gomega.BeNil())
	g.Expect(diff.Previous).NotTo(gomega.BeNil())
	g.Expect(diff.Previous.GetAnnotations()).To(gomega.BeNil())
	g.Expect(diff.Current.GetAnnotations()).To(gomega.HaveKey("team"))
}

func TestWriterSkipsUnchangedObjects(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	dir := t.TempDir()
	writer := NewWriter(config.OutputConfig{
		Directory: dir,
		Format:    config.OutputFormatYAML,
	})

	rule := config.ObjectRule{Kind: "ConfigMap"}
	obj := newUnstructured("v1", "ConfigMap", "default", "app-config")
	diff, err := writer.Process(rule, obj, nil)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(diff).NotTo(gomega.BeNil())

	diff, err = writer.Process(rule, obj, nil)
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(diff).To(gomega.BeNil())
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
	_, err := writer.Process(config.ObjectRule{Kind: "Node"}, obj, nil)

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
