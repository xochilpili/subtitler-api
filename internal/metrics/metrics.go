package metrics

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/xochilpili/subtitler-api/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

// InitMetrics configures the global MeterProvider and returns a shutdown func.
func InitMetrics(ctx context.Context, cfg *config.Config, logger *zerolog.Logger) func() {
	exporter, err := otlpmetricgrpc.New(ctx,
		otlpmetricgrpc.WithEndpoint(cfg.OtelEndpoint),
		otlpmetricgrpc.WithInsecure(),
	)
	if err != nil {
		logger.Fatal().Err(err).Msg("error while initializing otel metrics exporter")
	}

	res := resource.NewWithAttributes(
		semconv.SchemaURL,
		semconv.ServiceName(cfg.ServiceName),
		attribute.String("environment", cfg.ENV),
	)

	reader := sdkmetric.NewPeriodicReader(exporter) // default interval ~1m
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(reader),
		sdkmetric.WithResource(res),
	)
	otel.SetMeterProvider(mp)

	return func() {
		_ = mp.Shutdown(ctx)
	}
}
