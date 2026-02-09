package telemetry

import (
	"context"
	"fmt"

	"github.com/grafana/k8s-manifest-tail/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// MetricsRecorder defines the metric hooks emitted by the application.
type MetricsRecorder interface {
	RecordFullRun(ctx context.Context, count int)
	RecordManifestAdded(ctx context.Context)
	RecordManifestChanged(ctx context.Context)
	RecordManifestRemoved(ctx context.Context)
}

// Metrics exposes counters that are emitted for telemetry.
type Metrics struct {
	fullRunCounter metric.Int64Counter
	addedCounter   metric.Int64Counter
	changedCounter metric.Int64Counter
	removedCounter metric.Int64Counter
}

// RecordFullRun increments the manifest counter after a successful full run.
func (m *Metrics) RecordFullRun(ctx context.Context, count int) {
	if m == nil || m.fullRunCounter == nil {
		return
	}
	m.fullRunCounter.Add(ctx, int64(count))
}

// RecordManifestAdded increments the manifest addition counter.
func (m *Metrics) RecordManifestAdded(ctx context.Context) {
	if m.addedCounter != nil {
		m.addedCounter.Add(ctx, 1)
	}
}

// RecordManifestChanged increments the manifest change counter.
func (m *Metrics) RecordManifestChanged(ctx context.Context) {
	if m.changedCounter != nil {
		m.changedCounter.Add(ctx, 1)
	}
}

// RecordManifestRemoved increments the manifest removal counter.
func (m *Metrics) RecordManifestRemoved(ctx context.Context) {
	if m.removedCounter != nil {
		m.removedCounter.Add(ctx, 1)
	}
}

// SetupMetrics configures OpenTelemetry metric instruments.
func SetupMetrics(_ context.Context, _ config.LoggingConfig) (MetricsRecorder, func(context.Context) error, error) {
	meter := otel.GetMeterProvider().Meter("k8s-manifest-tail")
	fullRunCounter, err := meter.Int64Counter(
		"k8s_manifest_tail.full_run.manifests",
		metric.WithDescription("Number of manifests processed during a full refresh"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create full run counter: %w", err)
	}
	addedCounter, err := meter.Int64Counter(
		"k8s_manifest_tail.manifest.added",
		metric.WithDescription("Number of manifests written after being newly discovered"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create manifest added counter: %w", err)
	}
	changedCounter, err := meter.Int64Counter(
		"k8s_manifest_tail.manifest.changed",
		metric.WithDescription("Number of manifests updated due to detected changes"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create manifest changed counter: %w", err)
	}
	removedCounter, err := meter.Int64Counter(
		"k8s_manifest_tail.manifest.removed",
		metric.WithDescription("Number of manifests removed after deletion events"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create manifest removed counter: %w", err)
	}
	return &Metrics{
		fullRunCounter: fullRunCounter,
		addedCounter:   addedCounter,
		changedCounter: changedCounter,
		removedCounter: removedCounter,
	}, func(context.Context) error { return nil }, nil
}
