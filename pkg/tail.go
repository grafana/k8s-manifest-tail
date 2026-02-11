package pkg

import (
	"context"
	"errors"
	"fmt"
	"github.com/grafana/k8s-manifest-tail/internal/config"
	"github.com/grafana/k8s-manifest-tail/internal/discovery"
	"github.com/grafana/k8s-manifest-tail/internal/kube"
	"github.com/grafana/k8s-manifest-tail/internal/logging"
	"github.com/grafana/k8s-manifest-tail/internal/manifest"
	"github.com/grafana/k8s-manifest-tail/internal/telemetry"
	apierrors "k8s.io/apimachinery/pkg/api/errors"
	"k8s.io/apimachinery/pkg/api/meta"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/watch"
	"k8s.io/client-go/dynamic"
	"sync"
)

type Tail struct {
	Clients    *kube.Clients
	Config     *config.Config
	DiffLogger logging.DiffLogger
	Processor  manifest.Processor
	Metrics    telemetry.MetricsRecorder
}

func (t *Tail) RunFullManifestCheck(ctx context.Context) (int, error) {
	fetcher := discovery.NewFetcher(t.Clients, t.Config)
	var total int
	for _, rule := range t.Config.Objects {
		objects, err := fetcher.FetchResources(ctx, rule)
		if err != nil {
			return total, err
		}
		for i := range objects {
			obj := objects[i].DeepCopy()
			total++
			diff, err := t.Processor.Process(rule, obj, t.Config)
			if err != nil {
				return total, fmt.Errorf("process %s %s/%s: %w", rule.Kind, obj.GetNamespace(), obj.GetName(), err)
			}
			t.DiffLogger.Log(diff)
			t.recordDiffMetrics(ctx, diff)
		}
	}
	if t.Metrics != nil {
		t.Metrics.RecordFullRun(ctx, total)
	}
	return total, nil
}

func (t *Tail) WatchResources(ctx context.Context) error {
	errCh := make(chan error, len(t.Config.Objects))
	var wg sync.WaitGroup
	for _, rule := range t.Config.Objects {
		rule := rule
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := t.watchRule(ctx, rule); err != nil && !errors.Is(err, context.Canceled) {
				select {
				case errCh <- err:
				default:
				}
			}
		}()
	}

	done := make(chan struct{})
	go func() {
		wg.Wait()
		close(done)
	}()

	select {
	case err := <-errCh:
		return err
	case <-done:
		return ctx.Err()
	case <-ctx.Done():
		return ctx.Err()
	}
}

func (t *Tail) watchRule(ctx context.Context, rule config.ObjectRule) error {
	mapping, err := discovery.ResolveMapping(t.Clients.Mapper, rule)
	if err != nil {
		return err
	}

	resourceClient := t.Clients.Dynamic.Resource(mapping.Resource)
	exclude := discovery.NewExcludeSet(t.Config.ExcludeNamespaces)
	namespaces := discovery.EffectiveNamespaces(rule, t.Config)

	if mapping.Scope.Name() == meta.RESTScopeNameNamespace {
		if len(namespaces) == 0 {
			return t.watchResourceStream(ctx, resourceClient.Namespace(metav1.NamespaceAll), rule, exclude, true)
		}
		return t.watchNamespaceSet(ctx, resourceClient, rule, namespaces, exclude)
	}
	return t.watchResourceStream(ctx, resourceClient, rule, nil, false)
}

func (t *Tail) watchNamespaceSet(ctx context.Context, client dynamic.NamespaceableResourceInterface, rule config.ObjectRule, namespaces []string, exclude map[string]struct{}) error {
	errCh := make(chan error, len(namespaces))
	var wg sync.WaitGroup
	active := 0
	for _, ns := range namespaces {
		if discovery.ShouldExcludeNamespace(ns, exclude) {
			continue
		}
		active++
		nsClient := client.Namespace(ns)
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := t.watchResourceStream(ctx, nsClient, rule, nil, false); err != nil && !errors.Is(err, context.Canceled) {
				errCh <- err
			}
		}()
	}
	if active == 0 {
		return nil
	}
	go func() {
		wg.Wait()
		close(errCh)
	}()

	for err := range errCh {
		if err != nil {
			return err
		}
	}
	return nil
}

var errWatchClosed = errors.New("watch closed")

func (t *Tail) watchResourceStream(ctx context.Context, client dynamic.ResourceInterface, rule config.ObjectRule, exclude map[string]struct{}, filterExclude bool) error {
	for {
		if err := ctx.Err(); err != nil {
			return err
		}
		watcher, err := client.Watch(ctx, metav1.ListOptions{})
		if err != nil {
			if ctx.Err() != nil {
				return ctx.Err()
			}
			return fmt.Errorf("watch %s: %w", rule.Kind, err)
		}
		if err := t.consumeWatch(ctx, watcher, rule, exclude, filterExclude); err != nil {
			if errors.Is(err, errWatchClosed) {
				continue
			}
			return err
		}
		return nil
	}
}

func (t *Tail) consumeWatch(ctx context.Context, watcher watch.Interface, rule config.ObjectRule, exclude map[string]struct{}, filterExclude bool) error {
	result := watcher.ResultChan()
	for {
		select {
		case <-ctx.Done():
			watcher.Stop()
			return ctx.Err()
		case event, ok := <-result:
			if !ok {
				watcher.Stop()
				return errWatchClosed
			}
			obj, ok := event.Object.(*unstructured.Unstructured)
			if !ok {
				continue
			}
			if filterExclude && discovery.ShouldExcludeNamespace(obj.GetNamespace(), exclude) {
				continue
			}
			switch event.Type {
			case watch.Added, watch.Modified:
				diff, err := t.Processor.Process(rule, obj.DeepCopy(), t.Config)
				if err != nil {
					watcher.Stop()
					return fmt.Errorf("process %s %s/%s: %w", rule.Kind, obj.GetNamespace(), obj.GetName(), err)
				}
				t.DiffLogger.Log(diff)
				t.recordDiffMetrics(ctx, diff)
			case watch.Deleted:
				if err := t.Processor.Delete(rule, obj, t.Config); err != nil {
					watcher.Stop()
					return fmt.Errorf("delete %s %s/%s: %w", rule.Kind, obj.GetNamespace(), obj.GetName(), err)
				}
				diff := &manifest.Diff{Previous: obj}
				t.DiffLogger.Log(diff)
				t.recordDiffMetrics(ctx, diff)
			case watch.Error:
				watcher.Stop()
				return fmt.Errorf("watch error for %s: %v", rule.Kind, apierrors.FromObject(event.Object))
			}
		}
	}
}

func (t *Tail) recordDiffMetrics(ctx context.Context, diff *manifest.Diff) {
	if t.Metrics == nil || diff == nil {
		return
	}
	switch {
	case diff.Previous == nil && diff.Current != nil:
		t.Metrics.RecordManifestAdded(ctx)
	case diff.Previous != nil && diff.Current == nil:
		t.Metrics.RecordManifestRemoved(ctx)
	case diff.Previous != nil && diff.Current != nil:
		t.Metrics.RecordManifestChanged(ctx)
	}
}
