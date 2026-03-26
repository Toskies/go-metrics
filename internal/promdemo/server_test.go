package promdemo

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHandlerExposesMetrics(t *testing.T) {
	handler := NewHandler()

	workReq := httptest.NewRequest(http.MethodGet, "/work", nil)
	workRes := httptest.NewRecorder()
	handler.ServeHTTP(workRes, workReq)

	if workRes.Code != http.StatusOK {
		t.Fatalf("work status = %d, want %d", workRes.Code, http.StatusOK)
	}

	metricsReq := httptest.NewRequest(http.MethodGet, "/metrics", nil)
	metricsRes := httptest.NewRecorder()
	handler.ServeHTTP(metricsRes, metricsReq)

	body := metricsRes.Body.String()
	for _, want := range []string{
		"demo_http_requests_total",
		"demo_http_request_duration_seconds_bucket",
	} {
		if !strings.Contains(body, want) {
			t.Fatalf("metrics output missing %q\nbody:\n%s", want, body)
		}
	}
}
