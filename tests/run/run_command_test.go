package run_test

import (
	"fmt"
	"os"
	"path/filepath"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/grafana/k8s-manifest-tail/cmd"
	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/manifest"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("run command", func() {
	AfterEach(func() {
		cmd.SetKubeProvider(nil)
		cmd.SetManifestProcessor(nil)
		cmd.ResetConfiguration()
	})

	It("writes manifests to the configured directory", func() {
		outputDir := GinkgoT().TempDir()
		configPath := writeConfigFile(GinkgoT(), fmt.Sprintf(`
output:
  directory: %q
  format: yaml
objects:
  - apiVersion: v1
    kind: Pod
`, outputDir))
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"},
		}
		provider := newFakeProvider(
			[]runtime.Object{pod},
			[]resourceMapping{
				{
					GVR:   corev1.SchemeGroupVersion.WithResource("pods"),
					GVK:   corev1.SchemeGroupVersion.WithKind("Pod"),
					Scope: meta.RESTScopeNamespace,
				},
			},
		)
		cmd.SetKubeProvider(provider)

		stdout, stderr, err := runRunCommand(configPath)

		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("stderr: %s", stderr))
		Expect(stdout).To(ContainSubstring("Fetched 1 manifest"))
		path := filepath.Join(outputDir, "Pod", "default", "api.yaml")
		contents, readErr := os.ReadFile(path)
		Expect(readErr).NotTo(HaveOccurred())
		Expect(string(contents)).To(ContainSubstring("kind: Pod"))
	})

	It("processes fetched manifests", func() {
		configPath := writeConfigFile(GinkgoT(), `
output:
  directory: output
  format: yaml
objects:
  - apiVersion: v1
    kind: Pod
`)
		pod := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"},
		}
		provider := newFakeProvider(
			[]runtime.Object{pod},
			[]resourceMapping{
				{
					GVR:   corev1.SchemeGroupVersion.WithResource("pods"),
					GVK:   corev1.SchemeGroupVersion.WithKind("Pod"),
					Scope: meta.RESTScopeNamespace,
				},
			},
		)
		cmd.SetKubeProvider(provider)

		var processed []string
		cmd.SetManifestProcessor(&testProcessor{
			handler: func(rule config.ObjectRule, obj *unstructured.Unstructured, _ *config.Config) (*manifest.Diff, error) {
				processed = append(processed, fmt.Sprintf("%s/%s %s", obj.GetNamespace(), obj.GetName(), rule.Kind))
				return nil, nil
			},
		})

		stdout, stderr, err := runRunCommand(configPath)

		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("stderr: %s", stderr))
		Expect(stdout).To(ContainSubstring("Fetched 1 manifest"))
		Expect(processed).To(Equal([]string{"default/api Pod"}))
	})

	It("propagates processor errors", func() {
		configPath := writeConfigFile(GinkgoT(), `
objects:
  - apiVersion: v1
    kind: Pod
`)
		pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"}}
		provider := newFakeProvider(
			[]runtime.Object{pod},
			[]resourceMapping{
				{
					GVR:   corev1.SchemeGroupVersion.WithResource("pods"),
					GVK:   corev1.SchemeGroupVersion.WithKind("Pod"),
					Scope: meta.RESTScopeNamespace,
				},
			},
		)
		cmd.SetKubeProvider(provider)
		cmd.SetManifestProcessor(&testProcessor{
			handler: func(config.ObjectRule, *unstructured.Unstructured, *config.Config) (*manifest.Diff, error) {
				return nil, fmt.Errorf("write failed")
			},
		})

		_, stderr, err := runRunCommand(configPath)

		Expect(err).To(HaveOccurred())
		Expect(stderr).To(ContainSubstring("write failed"))
	})
})

type testProcessor struct {
	handler func(rule config.ObjectRule, obj *unstructured.Unstructured, cfg *config.Config) (*manifest.Diff, error)
}

func (t *testProcessor) Process(rule config.ObjectRule, obj *unstructured.Unstructured, cfg *config.Config) (*manifest.Diff, error) {
	if t.handler == nil {
		return nil, nil
	}
	return t.handler(rule, obj, cfg)
}
