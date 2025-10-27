package telemetry

import (
	"context"
	"errors"
	"log/slog"
	"time"

	"github.com/andrewhowdencom/vox/internal/config"
	"go.opentelemetry.io/contrib/instrumentation/runtime"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetrichttp"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/trace"
)

// Init initializes the OpenTelemetry SDK for both metrics and traces.
func Init(cfg *config.Config) (func(), error) {
	if cfg.Telemetry.OTLP.Endpoint == "" {
		slog.Debug("telemetry endpoint is not configured, skipping initialization")
		return func() {}, nil
	}

	slog.Info("initializing telemetry")
	ctx := context.Background()

	// Common HTTP options
	httpOpts := []otlptracehttp.Option{
		otlptracehttp.WithEndpoint(cfg.Telemetry.OTLP.Endpoint),
		otlptracehttp.WithTimeout(15 * time.Second),
	}
	if cfg.Telemetry.OTLP.Insecure {
		httpOpts = append(httpOpts, otlptracehttp.WithInsecure())
	}
	if len(cfg.Telemetry.OTLP.Headers) > 0 {
		httpOpts = append(httpOpts, otlptracehttp.WithHeaders(cfg.Telemetry.OTLP.Headers))
	}

	// Set up Trace Provider
	traceExporter, err := otlptracehttp.New(ctx, httpOpts...)
	if err != nil {
		return nil, err
	}
	traceProvider := trace.NewTracerProvider(
		trace.WithBatcher(traceExporter),
		trace.WithSampler(trace.AlwaysSample()),
	)
	otel.SetTracerProvider(traceProvider)

	// Set up Metric Provider
	// Note: The OTLP HTTP trace and metric clients have different option types, so we must re-create.
	metricOpts := []otlpmetrichttp.Option{
		otlpmetrichttp.WithEndpoint(cfg.Telemetry.OTLP.Endpoint),
		otlpmetrichttp.WithTimeout(15 * time.Second),
	}
	if cfg.Telemetry.OTLP.Insecure {
		metricOpts = append(metricOpts, otlpmetrichttp.WithInsecure())
	}
	if len(cfg.Telemetry.OTLP.Headers) > 0 {
		metricOpts = append(metricOpts, otlpmetrichttp.WithHeaders(cfg.Telemetry.OTLP.Headers))
	}
	metricExporter, err := otlpmetrichttp.New(ctx, metricOpts...)
	if err != nil {
		return nil, err
	}
	meterProvider := metric.NewMeterProvider(metric.WithReader(metric.NewPeriodicReader(metricExporter)))
	otel.SetMeterProvider(meterProvider)

	// Start runtime metrics collection
	if err := runtime.Start(); err != nil {
		return nil, err
	}

	slog.Info("telemetry initialized")

	// Create a single shutdown function for all providers
	shutdown := func() {
		slog.Info("shutting down telemetry")
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var shutdownErr error
		if err := meterProvider.Shutdown(ctx); err != nil {
			slog.Error("failed to shutdown meter provider", slog.Any("error", err))
			shutdownErr = errors.Join(shutdownErr, err)
		}
		if err := traceProvider.Shutdown(ctx); err != nil {
			slog.Error("failed to shutdown trace provider", slog.Any("error", err))
			shutdownErr = errors.Join(shutdownErr, err)
		}
		if shutdownErr != nil {
			slog.Error("telemetry shutdown failed", slog.Any("error", shutdownErr))
		}
	}

	return shutdown, nil
}
