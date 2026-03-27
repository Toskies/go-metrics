package otelstackdemo

import (
	"context"
	"errors"
	"net/http"
	"time"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/metric"
	"go.opentelemetry.io/otel/trace"
)

type Observability struct {
	tracer          trace.Tracer
	requestsTotal   metric.Int64Counter
	requestFailures metric.Int64Counter
	requestDuration metric.Float64Histogram
}

func NewObservability(tracer trace.Tracer, meter metric.Meter) (*Observability, error) {
	requestsTotal, err := meter.Int64Counter(
		"demo_http_requests_total",
		metric.WithDescription("Total HTTP requests served by the OTLP demo."),
	)
	if err != nil {
		return nil, err
	}

	requestFailures, err := meter.Int64Counter(
		"demo_http_request_failures_total",
		metric.WithDescription("Total failed HTTP requests served by the OTLP demo."),
	)
	if err != nil {
		return nil, err
	}

	requestDuration, err := meter.Float64Histogram(
		"demo_http_request_duration_seconds",
		metric.WithUnit("s"),
		metric.WithDescription("HTTP request latency for the OTLP demo."),
	)
	if err != nil {
		return nil, err
	}

	return &Observability{
		tracer:          tracer,
		requestsTotal:   requestsTotal,
		requestFailures: requestFailures,
		requestDuration: requestDuration,
	}, nil
}

func NewHandler(obs *Observability) http.Handler {
	mux := http.NewServeMux()
	mux.HandleFunc("/work", obs.instrument("/work", func(ctx context.Context, w http.ResponseWriter, r *http.Request) (int, error) {
		_, span := obs.tracer.Start(ctx, "simulate_work")
		defer span.End()

		time.Sleep(10 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("work complete"))
		return http.StatusOK, nil
	}))

	mux.HandleFunc("/checkout", obs.instrument("/checkout", func(ctx context.Context, w http.ResponseWriter, r *http.Request) (int, error) {
		_, span := obs.tracer.Start(ctx, "process_checkout")
		defer span.End()

		time.Sleep(15 * time.Millisecond)
		if r.URL.Query().Get("fail") == "1" {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("checkout failed"))
			return http.StatusInternalServerError, errors.New("checkout failed")
		}

		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("checkout ok"))
		return http.StatusOK, nil
	}))

	return mux
}

type appHandler func(ctx context.Context, w http.ResponseWriter, r *http.Request) (int, error)

func (o *Observability) instrument(route string, next appHandler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		ctx, span := o.tracer.Start(r.Context(), r.Method+" "+route)
		defer span.End()

		statusCode, err := next(ctx, w, r)
		attrs := metric.WithAttributes(
			attribute.String("http.request.method", r.Method),
			attribute.String("http.route", route),
			attribute.Int("http.response.status_code", statusCode),
		)

		o.requestsTotal.Add(ctx, 1, attrs)
		o.requestDuration.Record(ctx, time.Since(start).Seconds(), attrs)

		if err != nil {
			o.requestFailures.Add(ctx, 1, attrs)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			return
		}

		span.SetAttributes(
			attribute.String("http.request.method", r.Method),
			attribute.String("http.route", route),
			attribute.Int("http.response.status_code", statusCode),
		)
	}
}
