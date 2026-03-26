package promdemo

import (
	"net/http"
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func NewHandler() http.Handler {
	reg := prometheus.NewRegistry()

	requests := promauto.With(reg).NewCounterVec(
		prometheus.CounterOpts{
			Name: "demo_http_requests_total",
			Help: "Total HTTP requests served by the demo.",
		},
		[]string{"method", "route", "code"},
	)

	latency := promauto.With(reg).NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "demo_http_request_duration_seconds",
			Help:    "HTTP request latency for the demo.",
			Buckets: []float64{0.005, 0.01, 0.05, 0.1, 0.3, 1},
		},
		[]string{"method", "route"},
	)

	mux := http.NewServeMux()
	mux.Handle("/metrics", promhttp.HandlerFor(reg, promhttp.HandlerOpts{}))
	mux.HandleFunc("/work", func(w http.ResponseWriter, r *http.Request) {
		start := time.Now()
		time.Sleep(10 * time.Millisecond)

		statusCode := http.StatusOK
		w.WriteHeader(statusCode)
		_, _ = w.Write([]byte("prometheus client_golang demo"))

		requests.WithLabelValues(r.Method, "/work", strconv.Itoa(statusCode)).Inc()
		latency.WithLabelValues(r.Method, "/work").Observe(time.Since(start).Seconds())
	})

	return mux
}
