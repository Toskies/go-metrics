package oteldemo

import (
	"net/http"
	"time"

	promclient "github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.opentelemetry.io/otel/attribute"
	otelprom "go.opentelemetry.io/otel/exporters/prometheus"
	metricapi "go.opentelemetry.io/otel/metric"
	sdkmetric "go.opentelemetry.io/otel/sdk/metric"
)

func NewHandler() (http.Handler, error) {
	reg := promclient.NewRegistry()

	exporter, err := otelprom.New(otelprom.WithRegisterer(reg))
	if err != nil {
		return nil, err
	}

	provider := sdkmetric.NewMeterProvider(sdkmetric.WithReader(exporter))
	meter := provider.Meter("github.com/Toskies/go-metrics/internal/oteldemo")

	requests, err := meter.Int64Counter(
		"demo_http_requests",
		metricapi.WithDescription("Total HTTP requests served by the OpenTelemetry demo."),
	)
	if err != nil {
		return nil, err
	}

	latency, err := meter.Float64Histogram(
		"demo_http_request_duration_seconds",
		metricapi.WithUnit("s"),
		metricapi.WithDescription("HTTP request latency for the OpenTelemetry demo."),
	)
	if err != nil {
		return nil, err
	}

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		time.Sleep(10 * time.Millisecond)

		statusCode := http.StatusOK
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte("opentelemetry demo"))

		attrs := metricapi.WithAttributes(
			attribute.String("method", r.Method),
			attribute.String("route", "/work"),
			attribute.Int("code", statusCode),
		)

		requests.Add(r.Context(), 1, attrs)
		latency.Record(r.Context(), time.Since(start).Seconds(), attrs)
	})

	return mux, nil
}
