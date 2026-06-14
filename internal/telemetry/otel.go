package telemetry

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/prometheus"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.26.0"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// Setup initialises global OTel trace and metric providers.
// Returns a shutdown function the caller must defer.
// If otlpEndpoint is empty, trace export is disabled (useful in local dev).
func Setup(ctx context.Context, serviceName, version, environment, otlpEndpoint string) (shutdown func(context.Context) error, err error) {
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceName(serviceName),
			semconv.ServiceVersion(version),
			semconv.DeploymentEnvironment(environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("creating OTel resource: %w", err)
	}

	// ─── Metrics: Prometheus exporter ─────────────────────────────────────────
	promExporter, err := prometheus.New()
	if err != nil {
		return nil, fmt.Errorf("creating Prometheus exporter: %w", err)
	}
	meterProvider := metric.NewMeterProvider(
		metric.WithResource(res),
		metric.WithReader(promExporter),
	)
	otel.SetMeterProvider(meterProvider)

	// ─── Traces: OTLP gRPC exporter ───────────────────────────────────────────
	var traceProvider *sdktrace.TracerProvider
	if otlpEndpoint != "" {
		conn, connErr := grpc.NewClient(otlpEndpoint,
			grpc.WithTransportCredentials(insecure.NewCredentials()),
		)
		if connErr != nil {
			return nil, fmt.Errorf("creating gRPC connection to OTLP endpoint: %w", connErr)
		}

		traceExporter, traceErr := otlptracegrpc.New(ctx,
			otlptracegrpc.WithGRPCConn(conn),
		)
		if traceErr != nil {
			return nil, fmt.Errorf("creating OTLP trace exporter: %w", traceErr)
		}

		traceProvider = sdktrace.NewTracerProvider(
			sdktrace.WithResource(res),
			sdktrace.WithBatcher(traceExporter),
			sdktrace.WithSampler(sdktrace.AlwaysSample()),
		)
	} else {
		// No-op trace provider when OTLP endpoint is not configured
		traceProvider = sdktrace.NewTracerProvider(sdktrace.WithResource(res))
	}

	otel.SetTracerProvider(traceProvider)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return func(ctx context.Context) error {
		if err := traceProvider.Shutdown(ctx); err != nil {
			return err
		}
		return meterProvider.Shutdown(ctx)
	}, nil
}
