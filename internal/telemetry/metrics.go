package telemetry

import (
	"context"
	"fmt"

	"github.com/grafana/k8s-manifest-tail/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/metric"
)

// Metrics exposes counters that are emitted for telemetry.
type Metrics struct {
	fullRunCounter metric.Int64Counter
}

// RecordFullRun increments the manifest counter after a successful full run.
func (m *Metrics) RecordFullRun(ctx context.Context, count int) {
	if m == nil || m.fullRunCounter == nil {
		return
	}
	m.fullRunCounter.Add(ctx, int64(count))
}

// SetupMetrics configures OpenTelemetry metric instruments.
func SetupMetrics(_ context.Context, _ config.LoggingConfig) (*Metrics, func(context.Context) error, error) {
	meter := otel.GetMeterProvider().Meter("k8s-manifest-tail")
	counter, err := meter.Int64Counter(
		"manifests.total",
		metric.WithDescription("Number of manifests processed during a full refresh"),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("create full run counter: %w", err)
	}
	return &Metrics{fullRunCounter: counter}, func(context.Context) error { return nil }, nil
}
