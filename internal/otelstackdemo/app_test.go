package otelstackdemo

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"go.opentelemetry.io/otel/metric/noop"
	tracenoop "go.opentelemetry.io/otel/trace/noop"
)

func TestAppRoutes(t *testing.T) {
	obs, err := NewObservability(tracenoop.NewTracerProvider().Tracer("test"), noop.NewMeterProvider().Meter("test"))
	if err != nil {
		t.Fatalf("NewObservability() error = %v", err)
	}

	handler := NewHandler(obs)

	workReq := httptest.NewRequest(http.MethodGet, "/work", nil)
	workRes := httptest.NewRecorder()
	handler.ServeHTTP(workRes, workReq)
	if workRes.Code != http.StatusOK {
		t.Fatalf("/work status = %d, want %d", workRes.Code, http.StatusOK)
	}

	checkoutReq := httptest.NewRequest(http.MethodGet, "/checkout?fail=1", nil)
	checkoutRes := httptest.NewRecorder()
	handler.ServeHTTP(checkoutRes, checkoutReq)
	if checkoutRes.Code != http.StatusInternalServerError {
		t.Fatalf("/checkout?fail=1 status = %d, want %d", checkoutRes.Code, http.StatusInternalServerError)
	}
}
