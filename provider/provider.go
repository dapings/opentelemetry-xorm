package provider

import (
	"context"
	"log"

	runtimemetrics "go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	sdkresource "go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type OTELProvider interface {
	Shutdown(ctx context.Context) error
}

type defaultProvider struct {
	traceExp      sdktrace.SpanExporter
	metricsPusher sdkmetric.Exporter
}

func (p *defaultProvider) Shutdown(ctx context.Context) error {
	var err error

	if p.traceExp != nil {
		if err = p.traceExp.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
	}

	if p.metricsPusher != nil {
		if err = p.metricsPusher.Shutdown(ctx); err != nil {
			otel.Handle(err)
		}
	}

	return err
}

// NewOpenTelemetryProvider Initializes an otlp trace and metrics provider.
func NewOpenTelemetryProvider(opts ...Option) OTELProvider {
	var (
		err       error
		traceExp  sdktrace.SpanExporter
		metricExp sdkmetric.Exporter
	)

	ctx := context.TODO()

	cfg := newConfig(opts)

	if !cfg.enableTracing && !cfg.enableMetrics {
		return nil
	}

	// resource
	res := newResource(cfg)

	// propagator
	if otel.GetTextMapPropagator() == nil && cfg.textMapPropagator != nil {
		otel.SetTextMapPropagator(cfg.textMapPropagator)
	}

	// Tracing
	if cfg.enableTracing {
		// trace client
		var traceClientOpts []otlptracegrpc.Option
		if cfg.exportEndpoint != "" {
			traceClientOpts = append(traceClientOpts, otlptracegrpc.WithEndpoint(cfg.exportEndpoint))
		}
		if len(cfg.exportHeaders) > 0 {
			traceClientOpts = append(traceClientOpts, otlptracegrpc.WithHeaders(cfg.exportHeaders))
		}
		if cfg.exportInsecure {
			traceClientOpts = append(traceClientOpts, otlptracegrpc.WithInsecure())
		}

		traceClient := otlptracegrpc.NewClient(traceClientOpts...)

		// trace exporter
		traceExp, err = otlptrace.New(ctx, traceClient)
		if err != nil {
			log.Fatalf("failed to create otlp trace exporter: %s", err)
			return nil
		}

		// trace processor
		bsp := sdktrace.NewBatchSpanProcessor(traceExp)

		// trace provider
		tracerProvider := cfg.sdkTracerProvider
		if tracerProvider == nil {
			tracerProvider = sdktrace.NewTracerProvider(
				sdktrace.WithSampler(sdktrace.AlwaysSample()),
				sdktrace.WithResource(res),
				sdktrace.WithSpanProcessor(bsp),
			)
		}

		otel.SetTracerProvider(tracerProvider)
	}

	// Metrics
	if cfg.enableMetrics {
		metricsClientOpts := []otlpmetricgrpc.Option{
			otlpmetricgrpc.WithAggregationSelector(sdkmetric.DefaultAggregationSelector),
		}
		if cfg.exportEndpoint != "" {
			metricsClientOpts = append(metricsClientOpts, otlpmetricgrpc.WithEndpoint(cfg.exportEndpoint))
		}
		if len(cfg.exportHeaders) > 0 {
			metricsClientOpts = append(metricsClientOpts, otlpmetricgrpc.WithHeaders(cfg.exportHeaders))
		}
		if cfg.exportInsecure {
			metricsClientOpts = append(metricsClientOpts, otlpmetricgrpc.WithInsecure())
		}

		// metrics exporter
		metricExp, err = otlpmetricgrpc.New(ctx,
			metricsClientOpts...,
		)
		handleInitErr(err, "Failed to create the collector metric exporter")

		// metrics pusher
		meterProvider := sdkmetric.NewMeterProvider(
			sdkmetric.WithReader(sdkmetric.NewPeriodicReader(metricExp)),
		)
		otel.SetMeterProvider(meterProvider)

		err = runtimemetrics.Start()
		handleInitErr(err, "Failed to start runtime metrics collector")
	}

	return &defaultProvider{
		traceExp:      traceExp,
		metricsPusher: metricExp,
	}
}

func newResource(cfg *config) *sdkresource.Resource {
	if cfg.resource != nil {
		return cfg.resource
	}

	ctx := context.Background()
	res, err := sdkresource.New(ctx,
		sdkresource.WithFromEnv(),
		sdkresource.WithProcess(),
		sdkresource.WithProcessPID(),
		sdkresource.WithTelemetrySDK(),
		sdkresource.WithHost(),
		sdkresource.WithOS(),
		sdkresource.WithContainer(),
		sdkresource.WithProcessRuntimeName(),
		sdkresource.WithSchemaURL(semconv.SchemaURL),
		sdkresource.WithDetectors(cfg.resourceDetectors...),
		sdkresource.WithAttributes(cfg.resourceAttributes...),
	)

	if err != nil {
		otel.Handle(err)
		return sdkresource.Default()
	}

	return res
}

func handleInitErr(err error, msg string) {
	if err != nil {
		log.Fatalf("%s: %+v", msg, err)
	}
}
