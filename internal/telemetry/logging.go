package telemetry

import (
	"context"
	"fmt"
	"github.com/grafana/k8s-manifest-tail/internal"
	"github.com/grafana/k8s-manifest-tail/internal/config"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploggrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlplog/otlploghttp"
	"go.opentelemetry.io/otel/exporters/stdout/stdoutlog"
	"go.opentelemetry.io/otel/log"
	sdklog "go.opentelemetry.io/otel/sdk/log"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.24.0"
	"os"
)

// SetupLogging configures an OTLP log exporter if enabled. It returns the logger, a shutdown func, and any error.
func SetupLogging(ctx context.Context, cfg config.LoggingConfig) (log.Logger, func(context.Context) error, error) {
	consoleExporter, err := stdoutlog.New(stdoutlog.WithWriter(os.Stdout))
	if err != nil {
		return nil, nil, fmt.Errorf("create console log exporter: %w", err)
	}
	res, err := resource.Merge(
		resource.Default(),
		resource.NewSchemaless(
			semconv.ServiceNameKey.String("k8s-manifest-tail"),
			semconv.ServiceVersionKey.String(internal.Version),
		),
	)
	if err != nil {
		return nil, nil, fmt.Errorf("describe service resource: %w", err)
	}

	var processors []sdklog.Processor
	processors = append(processors, sdklog.NewSimpleProcessor(consoleExporter))

	if cfg.OTLP.Enabled() {
		exporter, err := newOTLPExporter(ctx, cfg.OTLP)
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

func Info(logger log.Logger, msg string, attributes ...log.KeyValue) {
	var record log.Record
	record.SetSeverity(log.SeverityInfo)
	record.SetBody(log.StringValue(msg))
	record.AddAttributes(attributes...)
	logger.Emit(context.Background(), record)
}

const (
	OTLPProtocolGRPC         string = "grpc"
	OTLPProtocolHTTPJSON     string = "http/json"
	OTLPProtocolHTTPProtobuf string = "http/protobuf"
)

func newOTLPExporter(ctx context.Context, cfg config.OTLPConfig) (sdklog.Exporter, error) {
	protocol := OTLPProtocolGRPC
	if os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL") != "" {
		protocol = os.Getenv("OTEL_EXPORTER_OTLP_LOGS_PROTOCOL")
	} else if os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL") != "" {
		protocol = os.Getenv("OTEL_EXPORTER_OTLP_PROTOCOL")
	}

	switch protocol {
	case OTLPProtocolGRPC:
		opts := []otlploggrpc.Option{}
		if cfg.Endpoint != "" {
			opts = append(opts, otlploggrpc.WithEndpoint(cfg.Endpoint))
		}
		if cfg.Insecure {
			opts = append(opts, otlploggrpc.WithInsecure())
		}
		return otlploggrpc.New(ctx, opts...)
	case OTLPProtocolHTTPJSON, OTLPProtocolHTTPProtobuf:
		opts := []otlploghttp.Option{}
		if cfg.Endpoint != "" {
			opts = append(opts, otlploghttp.WithEndpoint(cfg.Endpoint))
		}
		if cfg.Insecure {
			opts = append(opts, otlploghttp.WithInsecure())
		}
		return otlploghttp.New(ctx, opts...)
	default:
		return nil, fmt.Errorf("unknown OTLP protocol: %s", protocol)
	}
}
