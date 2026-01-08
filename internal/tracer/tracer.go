package tracer

import (
	"context"

	"github.com/rs/zerolog"
	"github.com/xochilpili/subtitler-api/internal/config"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/sdk/resource"
	"go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
)

type Tracer struct {
	config *config.Config
	logger *zerolog.Logger
}

func InitTracer(ctx context.Context, config *config.Config, logger *zerolog.Logger) func() {
	exporter, err := otlptracegrpc.New(ctx, otlptracegrpc.WithEndpoint(config.OtelEndpoint), otlptracegrpc.WithInsecure())
	if err != nil {
		logger.Fatal().Err(err).Msg("error while initializing otel exporter")
	}

	tp := trace.NewTracerProvider(trace.WithBatcher(exporter), trace.WithResource(resource.NewWithAttributes(semconv.SchemaURL, semconv.ServiceName(config.ServiceName), attribute.String("environment", config.ENV))))
	otel.SetTracerProvider(tp)

	return func() {
		_ = tp.Shutdown(ctx)
	}
}
