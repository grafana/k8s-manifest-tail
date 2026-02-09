package discovery

import (
	"context"
	"testing"

	"github.com/onsi/gomega"

	appsv1 "k8s.io/api/apps/v1"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	dynamicfake "k8s.io/client-go/dynamic/fake"
	clientgoscheme "k8s.io/client-go/kubernetes/scheme"

	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/kube"
)

var testScheme = runtime.NewScheme()

func init() {
	_ = clientgoscheme.AddToScheme(testScheme)
}

func TestFetcherListsPodsAcrossNamespaces(t *testing.T) {
	g := gomega.NewWithT(t)
	ctx := context.Background()
	cfg := &config.Config{}
	podDefault := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"},
	}
	podProd := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "api-prod", Namespace: "prod"},
	}
	clients := newTestClients(
		[]runtime.Object{podDefault, podProd},
		[]resourceMapping{
			{
				GVR:   corev1.SchemeGroupVersion.WithResource("pods"),
				GVK:   corev1.SchemeGroupVersion.WithKind("Pod"),
				Scope: meta.RESTScopeNamespace,
			},
		},
	)

	fetcher := NewFetcher(clients, cfg)
	items, err := fetcher.FetchResources(ctx, config.ObjectRule{APIVersion: "v1", Kind: "Pod"})

	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(objectNames(items)).To(gomega.ContainElements("api", "api-prod"))
}

func TestFetcherRespectsGlobalNamespaces(t *testing.T) {
	g := gomega.NewWithT(t)
	ctx := context.Background()
	cfg := &config.Config{
		Namespaces: []string{"default"},
	}
	podDefault := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "db", Namespace: "default"}}
	podProd := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "db-prod", Namespace: "prod"}}
	clients := newTestClients(
		[]runtime.Object{podDefault, podProd},
		[]resourceMapping{
			{
				GVR:   corev1.SchemeGroupVersion.WithResource("pods"),
				GVK:   corev1.SchemeGroupVersion.WithKind("Pod"),
				Scope: meta.RESTScopeNamespace,
			},
		},
	)

	fetcher := NewFetcher(clients, cfg)
	items, err := fetcher.FetchResources(ctx, config.ObjectRule{APIVersion: "v1", Kind: "Pod"})

	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(objectNames(items)).To(gomega.Equal([]string{"db"}))
}

func TestFetcherRespectsRuleNamespaceOverrides(t *testing.T) {
	g := gomega.NewWithT(t)
	ctx := context.Background()
	cfg := &config.Config{
		Namespaces: []string{"prod"},
	}
	deployDefault := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "frontend", Namespace: "default"}}
	deployProd := &appsv1.Deployment{ObjectMeta: metav1.ObjectMeta{Name: "frontend", Namespace: "prod"}}
	clients := newTestClients(
		[]runtime.Object{deployDefault, deployProd},
		[]resourceMapping{
			{
				GVR:   appsv1.SchemeGroupVersion.WithResource("deployments"),
				GVK:   appsv1.SchemeGroupVersion.WithKind("Deployment"),
				Scope: meta.RESTScopeNamespace,
			},
		},
	)

	fetcher := NewFetcher(clients, cfg)
	items, err := fetcher.FetchResources(ctx, config.ObjectRule{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Namespaces: []string{"default"},
	})

	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(objectNames(items)).To(gomega.Equal([]string{"frontend"}))
}

func TestFetcherAppliesExcludeNamespaces(t *testing.T) {
	g := gomega.NewWithT(t)
	ctx := context.Background()
	cfg := &config.Config{
		ExcludeNamespaces: []string{"kube-system"},
	}
	systemPod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "coredns", Namespace: "kube-system"}}
	userPod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "worker", Namespace: "default"}}
	clients := newTestClients(
		[]runtime.Object{systemPod, userPod},
		[]resourceMapping{
			{
				GVR:   corev1.SchemeGroupVersion.WithResource("pods"),
				GVK:   corev1.SchemeGroupVersion.WithKind("Pod"),
				Scope: meta.RESTScopeNamespace,
			},
		},
	)

	fetcher := NewFetcher(clients, cfg)
	items, err := fetcher.FetchResources(ctx, config.ObjectRule{APIVersion: "v1", Kind: "Pod"})

	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(objectNames(items)).To(gomega.Equal([]string{"worker"}))
}

func TestFetcherHandlesClusterScopedResources(t *testing.T) {
	g := gomega.NewWithT(t)
	ctx := context.Background()
	cfg := &config.Config{}
	node := &corev1.Node{ObjectMeta: metav1.ObjectMeta{Name: "node-a"}}
	clients := newTestClients(
		[]runtime.Object{node},
		[]resourceMapping{
			{
				GVR:   corev1.SchemeGroupVersion.WithResource("nodes"),
				GVK:   corev1.SchemeGroupVersion.WithKind("Node"),
				Scope: meta.RESTScopeRoot,
			},
		},
	)

	fetcher := NewFetcher(clients, cfg)
	items, err := fetcher.FetchResources(ctx, config.ObjectRule{APIVersion: "v1", Kind: "Node"})

	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(objectNames(items)).To(gomega.Equal([]string{"node-a"}))
}

func TestFetcherAppliesNamePattern(t *testing.T) {
	g := gomega.NewWithT(t)
	ctx := context.Background()
	cfg := &config.Config{}
	alloyLogs := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "alloy-logs", Namespace: "default"}}
	alloyMetrics := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "alloy-metrics", Namespace: "default"}}
	nodeExporter := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "nodeexporter", Namespace: "default"}}
	clients := newTestClients(
		[]runtime.Object{alloyLogs, alloyMetrics, nodeExporter},
		[]resourceMapping{
			{
				GVR:   corev1.SchemeGroupVersion.WithResource("pods"),
				GVK:   corev1.SchemeGroupVersion.WithKind("Pod"),
				Scope: meta.RESTScopeNamespace,
			},
		},
	)

	fetcher := NewFetcher(clients, cfg)
	items, err := fetcher.FetchResources(ctx, config.ObjectRule{
		APIVersion:  "v1",
		Kind:        "Pod",
		NamePattern: "alloy-.*",
	})

	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(objectNames(items)).To(gomega.Equal([]string{"alloy-logs", "alloy-metrics"}))
}

func objectNames(items []unstructured.Unstructured) []string {
	var names []string
	for _, item := range items {
		names = append(names, item.GetName())
	}
	return names
}

type resourceMapping struct {
	GVR   schema.GroupVersionResource
	GVK   schema.GroupVersionKind
	Scope meta.RESTScope
}

func newTestClients(objects []runtime.Object, mappings []resourceMapping) *kube.Clients {
	dynamicClient := dynamicfake.NewSimpleDynamicClient(testScheme, objects...)
	mapper := newRESTMapper(mappings)
	return &kube.Clients{
		Dynamic: dynamicClient,
		Mapper:  mapper,
	}
}

func newRESTMapper(mappings []resourceMapping) meta.RESTMapper {
	groupVersions := make(map[schema.GroupVersion]struct{})
	for _, m := range mappings {
		gv := schema.GroupVersion{Group: m.GVR.Group, Version: m.GVR.Version}
		groupVersions[gv] = struct{}{}
	}
	var gvList []schema.GroupVersion
	for gv := range groupVersions {
		gvList = append(gvList, gv)
	}
	mapper := meta.NewDefaultRESTMapper(gvList)
	for _, m := range mappings {
		mapper.AddSpecific(m.GVK, m.GVR, m.GVR, m.Scope)
	}
	return mapper
}
