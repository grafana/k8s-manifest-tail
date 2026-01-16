package list_test

import (
	"fmt"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"

	"github.com/grafana/k8s-manifest-tail/cmd"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime"
)

var _ = Describe("list command", func() {
	AfterEach(func() {
		cmd.SetKubeProvider(nil)
		cmd.ResetConfiguration()
	})

	It("lists pods across namespaces", func() {
		configPath := writeConfigFile(GinkgoT(), `
output:
  directory: output
  format: yaml
objects:
  - apiVersion: v1
    kind: Pod
`)
		podDefault := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"},
		}
		podProd := &corev1.Pod{
			ObjectMeta: metav1.ObjectMeta{Name: "api-prod", Namespace: "prod"},
		}
		provider := newFakeProvider(
			[]runtime.Object{podDefault, podProd},
			[]resourceMapping{
				{
					GVR:   corev1.SchemeGroupVersion.WithResource("pods"),
					GVK:   corev1.SchemeGroupVersion.WithKind("Pod"),
					Scope: meta.RESTScopeNamespace,
				},
			},
		)
		cmd.SetKubeProvider(provider)

		stdout, stderr, err := runListCommand(configPath)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("stderr: %s", stderr))
		Expect(stdout).To(ContainSubstring("Pod"))
		Expect(stdout).To(ContainSubstring("api"))
		Expect(stdout).To(ContainSubstring("api-prod"))
	})

	It("respects global namespace filters", func() {
		configPath := writeConfigFile(GinkgoT(), `
namespaces: ["default"]
objects:
  - apiVersion: v1
    kind: Pod
`)
		podDefault := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "db", Namespace: "default"}}
		podProd := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "db-prod", Namespace: "prod"}}
		provider := newFakeProvider(
			[]runtime.Object{podDefault, podProd},
			[]resourceMapping{
				{
					GVR:   corev1.SchemeGroupVersion.WithResource("pods"),
					GVK:   corev1.SchemeGroupVersion.WithKind("Pod"),
					Scope: meta.RESTScopeNamespace,
				},
			},
		)
		cmd.SetKubeProvider(provider)

		stdout, stderr, err := runListCommand(configPath)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("stderr: %s", stderr))
		Expect(stdout).To(ContainSubstring("db"))
		Expect(stdout).NotTo(ContainSubstring("db-prod"))
	})

	It("applies object-level namespace overrides", func() {
		configPath := writeConfigFile(GinkgoT(), `
objects:
  - apiVersion: apps/v1
    kind: Deployment
    namespaces: ["default"]
`)
		deployDefault := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "frontend", Namespace: "default"},
		}
		deployProd := &appsv1.Deployment{
			ObjectMeta: metav1.ObjectMeta{Name: "frontend", Namespace: "prod"},
		}
		provider := newFakeProvider(
			[]runtime.Object{deployDefault, deployProd},
			[]resourceMapping{
				{
					GVR:   appsv1.SchemeGroupVersion.WithResource("deployments"),
					GVK:   appsv1.SchemeGroupVersion.WithKind("Deployment"),
					Scope: meta.RESTScopeNamespace,
				},
			},
		)
		cmd.SetKubeProvider(provider)

		stdout, stderr, err := runListCommand(configPath)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("stderr: %s", stderr))
		Expect(stdout).To(ContainSubstring("frontend"))
		Expect(stdout).NotTo(ContainSubstring("prod"))
	})

	It("excludes namespaces from the results", func() {
		configPath := writeConfigFile(GinkgoT(), `
excludeNamespaces: ["kube-system"]
objects:
  - apiVersion: v1
    kind: Pod
`)
		systemPod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "coredns", Namespace: "kube-system"}}
		userPod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "worker", Namespace: "default"}}
		provider := newFakeProvider(
			[]runtime.Object{systemPod, userPod},
			[]resourceMapping{
				{
					GVR:   corev1.SchemeGroupVersion.WithResource("pods"),
					GVK:   corev1.SchemeGroupVersion.WithKind("Pod"),
					Scope: meta.RESTScopeNamespace,
				},
			},
		)
		cmd.SetKubeProvider(provider)

		stdout, stderr, err := runListCommand(configPath)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("stderr: %s", stderr))
		Expect(stdout).To(ContainSubstring("worker"))
		Expect(stdout).NotTo(ContainSubstring("coredns"))
	})

	It("prints friendly message when nothing matches", func() {
		configPath := writeConfigFile(GinkgoT(), `
objects:
  - apiVersion: v1
    kind: Pod
`)
		provider := newFakeProvider(
			nil,
			[]resourceMapping{
				{
					GVR:   corev1.SchemeGroupVersion.WithResource("pods"),
					GVK:   corev1.SchemeGroupVersion.WithKind("Pod"),
					Scope: meta.RESTScopeNamespace,
				},
			},
		)
		cmd.SetKubeProvider(provider)

		stdout, stderr, err := runListCommand(configPath)
		Expect(err).NotTo(HaveOccurred(), fmt.Sprintf("stderr: %s", stderr))
		Expect(stdout).To(ContainSubstring("No resources found"))
	})
})
