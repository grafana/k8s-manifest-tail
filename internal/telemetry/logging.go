package telemetry

import (
	"context"
	"fmt"
	"os"

	"github.com/grafana/k8s-manifest-tail/internal/config"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	otellog "go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
)

// SetupLogging configures an OTLP log exporter if enabled. It returns the logger, a shutdown func, and any error.
func SetupLogging(ctx context.Context, cfg config.LoggingConfig) (otellog.Logger, func(context.Context) error, error) {
	consoleExporter, err := stdoutlog.New(stdoutlog.WithWriter(os.Stdout))
	if err != nil {
		return nil, nil, fmt.Errorf("create console log exporter: %w", err)
	}
	opts := []otlploggrpc.Option{}
	if cfg.OTLP.Endpoint != "" {
		opts = append(opts, otlploggrpc.WithEndpoint(cfg.OTLP.Endpoint))
	}
	if cfg.OTLP.Insecure {
		opts = append(opts, otlploggrpc.WithInsecure())
	}

	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceNameKey.String("k8s-manifest-tail"),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("describe service resource: %w", err)
	}

	var processors []sdklog.Processor
	processors = append(processors, sdklog.NewSimpleProcessor(consoleExporter))

	if cfg.OTLP.ShouldEnable() {
		exporter, err := otlploggrpc.New(ctx, opts...)
		if err != nil {
			return nil, nil, fmt.Errorf("create otlp log exporter: %w", err)
		}
		processors = append(processors, sdklog.NewBatchProcessor(exporter))
	}

	providerOpts := []sdklog.LoggerProviderOption{sdklog.WithResource(res)}
	for _, proc := range processors {
		providerOpts = append(providerOpts, sdklog.WithProcessor(proc))
	}
	provider := sdklog.NewLoggerProvider(providerOpts...)

	logger := provider.Logger("k8s-manifest-tail")
	shutdown := func(ctx context.Context) error {
		return provider.Shutdown(ctx)
	}
	return logger, shutdown, nil
}

func Info(logger otellog.Logger, msg string, attributes ...otellog.KeyValue) {
	var record otellog.Record
	record.SetSeverity(otellog.SeverityInfo)
	record.SetBody(otellog.StringValue(msg))
	record.AddAttributes(attributes...)
	logger.Emit(context.Background(), record)
}
