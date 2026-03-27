package otelstackdemo

import (
	"context"
	"errors"
	"fmt"
	"net"
	"os"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/otlp/otlpmetric/otlpmetricgrpc"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/propagation"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	"go.opentelemetry.io/otel/trace"
)

type Config struct {
	ServiceName  string
	ServiceVer   string
	Environment  string
	OTLPEndpoint string
	Insecure     bool
}

func (c Config) Validate() error {
	if c.ServiceName == "" {
		return errors.New("service name is required")
	}
	if c.OTLPEndpoint == "" {
		return errors.New("otlp endpoint is required")
	}
	if _, _, err := net.SplitHostPort(c.OTLPEndpoint); err != nil {
		return fmt.Errorf("invalid otlp endpoint %q: %w", c.OTLPEndpoint, err)
	}
	return nil
}

func DefaultConfig() Config {
	return Config{
		ServiceName:  getenv("OTEL_SERVICE_NAME", "otelstackdemo"),
		ServiceVer:   getenv("OTEL_SERVICE_VERSION", "dev"),
		Environment:  getenv("OTEL_DEPLOYMENT_ENVIRONMENT", "local"),
		OTLPEndpoint: getenv("OTEL_EXPORTER_OTLP_ENDPOINT", "localhost:4317"),
		Insecure:     true,
	}
}

type Providers struct {
	TracerProvider *sdktrace.TracerProvider
	MeterProvider  *sdkmetric.MeterProvider
}

func Setup(ctx context.Context, cfg Config) (*Providers, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			attribute.String("service.name", cfg.ServiceName),
			attribute.String("service.version", cfg.ServiceVer),
			attribute.String("deployment.environment", cfg.Environment),
		),
	)
	if err != nil {
		return nil, err
	}

	traceOptions := []otlptracegrpc.Option{
		otlptracegrpc.WithEndpoint(cfg.OTLPEndpoint),
	}
	metricOptions := []otlpmetricgrpc.Option{
		otlpmetricgrpc.WithEndpoint(cfg.OTLPEndpoint),
	}
	if cfg.Insecure {
		traceOptions = append(traceOptions, otlptracegrpc.WithInsecure())
		metricOptions = append(metricOptions, otlpmetricgrpc.WithInsecure())
	}

	traceExporter, err := otlptracegrpc.New(ctx, traceOptions...)
	if err != nil {
		return nil, err
	}

	metricExporter, err := otlpmetricgrpc.New(ctx, metricOptions...)
	if err != nil {
		return nil, err
	}

	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(traceExporter),
		sdktrace.WithResource(res),
	)
	mp := sdkmetric.NewMeterProvider(
		sdkmetric.WithReader(
			sdkmetric.NewPeriodicReader(metricExporter, sdkmetric.WithInterval(5*time.Second)),
		),
		sdkmetric.WithResource(res),
	)

	otel.SetTracerProvider(tp)
	otel.SetMeterProvider(mp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	return &Providers{
		TracerProvider: tp,
		MeterProvider:  mp,
	}, nil
}

func (p *Providers) Shutdown(ctx context.Context) error {
	if err := p.MeterProvider.Shutdown(ctx); err != nil {
		return err
	}
	return p.TracerProvider.Shutdown(ctx)
}

func (p *Providers) Tracer(name string) trace.Tracer {
	return p.TracerProvider.Tracer(name)
}

func (p *Providers) Meter(name string) metric.Meter {
	return p.MeterProvider.Meter(name)
}

func getenv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
