package otelstackdemo

import "testing"

func TestTelemetryConfig(t *testing.T) {
	cfg := Config{
		ServiceName:  "checkout-service",
		OTLPEndpoint: "",
		Insecure:     true,
	}

	if err := cfg.Validate(); err == nil {
		t.Fatal("Validate() error = nil, want non-nil for missing OTLP endpoint")
	}
}
