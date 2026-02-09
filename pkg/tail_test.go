package pkg

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/onsi/gomega"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic/fake"

	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/kube"
	"github.com/grafana/k8s-manifest-tail/internal/manifest"
)

var testScheme = runtime.NewScheme()

func init() {
	_ = corev1.AddToScheme(testScheme)
}

func TestTailRunFullManifestCheck(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	pod := &corev1.Pod{ObjectMeta: metav1.ObjectMeta{Name: "api", Namespace: "default"}}
	dyn := fake.NewSimpleDynamicClient(testScheme, pod)
	mapper := newRESTMapper([]resourceMapping{{
		GVR:   corev1.SchemeGroupVersion.WithResource("pods"),
		GVK:   corev1.SchemeGroupVersion.WithKind("Pod"),
		Scope: meta.RESTScopeNamespace,
	}})

	stubProc := &stubProcessor{}
	stubLogger := &stubDiffLogger{}

	tail := Tail{
		Clients: &kube.Clients{
			Dynamic: dyn,
			Mapper:  mapper,
		},
		Config: &config.Config{
			Objects: []config.ObjectRule{
				{APIVersion: "v1", Kind: "Pod"},
			},
		},
		DiffLogger: stubLogger,
		Processor:  stubProc,
	}

	total, err := tail.RunFullManifestCheck(context.Background())
	g.Expect(err).NotTo(gomega.HaveOccurred())
	g.Expect(total).To(gomega.Equal(1))
	g.Expect(stubProc.processed).To(gomega.Equal([]string{"default/api"}))
	g.Expect(stubLogger.logged).To(gomega.Equal(1))
}

func TestTailConsumeWatchHandlesEvents(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	stubProc := &stubProcessor{}
	stubLogger := &stubDiffLogger{}

	tail := Tail{
		Config: &config.Config{
			Objects: []config.ObjectRule{{APIVersion: "v1", Kind: "Pod"}},
		},
		Processor:  stubProc,
		DiffLogger: stubLogger,
	}

	watcher := watch.NewFake()
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	obj := &unstructured.Unstructured{}
	obj.SetAPIVersion("v1")
	obj.SetKind("Pod")
	obj.SetNamespace("default")
	obj.SetName("api")

	var wg sync.WaitGroup
	wg.Add(1)
	go func() {
		defer wg.Done()
		time.Sleep(10 * time.Millisecond)
		watcher.Add(obj)
		watcher.Delete(obj)
		watcher.Stop()
	}()

	err := tail.consumeWatch(ctx, watcher, config.ObjectRule{APIVersion: "v1", Kind: "Pod"}, nil, false)
	g.Expect(err).To(gomega.Equal(errWatchClosed))
	wg.Wait()

	g.Expect(stubProc.processed).To(gomega.Equal([]string{"default/api"}))
	g.Expect(stubProc.deleted).To(gomega.Equal([]string{"default/api"}))
	g.Expect(stubLogger.logged).To(gomega.Equal(2))
}

func TestTailRecordDiffMetrics(t *testing.T) {
	t.Parallel()
	g := gomega.NewWithT(t)

	ctx := context.Background()
	metrics := &stubMetrics{}
	tail := Tail{Metrics: metrics}

	addObj := &unstructured.Unstructured{}
	changeObjPrev := &unstructured.Unstructured{}
	changeObjCurr := &unstructured.Unstructured{}

	tail.recordDiffMetrics(ctx, &manifest.Diff{Current: addObj})
	tail.recordDiffMetrics(ctx, &manifest.Diff{Previous: changeObjPrev, Current: changeObjCurr})
	tail.recordDiffMetrics(ctx, &manifest.Diff{Previous: addObj})
	tail.recordDiffMetrics(ctx, nil) // ignored

	g.Expect(metrics.added).To(gomega.Equal(1))
	g.Expect(metrics.changed).To(gomega.Equal(1))
	g.Expect(metrics.removed).To(gomega.Equal(1))
}

type stubProcessor struct {
	processed []string
	deleted   []string
}

func (s *stubProcessor) Process(_ config.ObjectRule, obj *unstructured.Unstructured, _ *config.Config) (*manifest.Diff, error) {
	s.processed = append(s.processed, fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName()))
	return &manifest.Diff{Current: obj}, nil
}

func (s *stubProcessor) Delete(_ config.ObjectRule, obj *unstructured.Unstructured, _ *config.Config) error {
	s.deleted = append(s.deleted, fmt.Sprintf("%s/%s", obj.GetNamespace(), obj.GetName()))
	return nil
}

type stubDiffLogger struct {
	logged int
}

func (s *stubDiffLogger) Log(_ *manifest.Diff) {
	s.logged++
}

type stubMetrics struct {
	fullRuns []int
	added    int
	changed  int
	removed  int
}

func (s *stubMetrics) RecordFullRun(_ context.Context, count int) {
	s.fullRuns = append(s.fullRuns, count)
}

func (s *stubMetrics) RecordManifestAdded(context.Context) {
	s.added++
}

func (s *stubMetrics) RecordManifestChanged(context.Context) {
	s.changed++
}

func (s *stubMetrics) RecordManifestRemoved(context.Context) {
	s.removed++
}

type resourceMapping struct {
	GVR   schema.GroupVersionResource
	GVK   schema.GroupVersionKind
	Scope meta.RESTScope
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
