package tracing

import (
	"context"
	"log"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.4.0"
)

type Option func(*config)

type config struct {
    samplingRatio float64
}

func WithSamplingRatio(ratio float64) Option {
    return func(c *config) {
        c.samplingRatio = ratio
    }
}

// InitTracer initializes the OpenTelemetry tracer
func InitTracer(serviceName string, collectorURL string, opts ...Option) (func(), error) {
    cfg := &config{
        samplingRatio: 1.0,
    }

    for _, opt := range opts {
        opt(cfg)
    }

    exporter, err := otlptracehttp.New(
        context.Background(),
        otlptracehttp.WithEndpoint(collectorURL),
        otlptracehttp.WithInsecure(),
    )
    if err != nil {
        return nil, err
    }

    res, err := resource.New(
        context.Background(),
        resource.WithAttributes(
            semconv.ServiceNameKey.String(serviceName),
        ),
    )
    if err != nil {
        return nil, err
    }

    provider := sdktrace.NewTracerProvider(
        sdktrace.WithSampler(sdktrace.TraceIDRatioBased(cfg.samplingRatio)),
        sdktrace.WithBatcher(exporter),
        sdktrace.WithResource(res),
    )

    otel.SetTracerProvider(provider)

    return func() {
        if err := provider.Shutdown(context.Background()); err != nil {
            log.Printf("Error shutting down tracer provider: %v", err)
        }
    }, nil
}
