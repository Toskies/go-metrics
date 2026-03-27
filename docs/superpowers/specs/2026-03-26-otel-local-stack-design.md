# OTel Local Stack Design

## Goal

Add a complete local OpenTelemetry example that demonstrates both metrics and traces from a Go service without exposing an application `/metrics` endpoint. The example should run with the Go app on the host machine and the observability stack in Docker Compose.

## Scope

This design covers:

- a new Go demo application that exports traces and metrics over OTLP/gRPC
- a local stack with OpenTelemetry Collector, Prometheus, Tempo, and Grafana
- configuration and docs required to run the stack locally end-to-end

This design intentionally does not cover:

- logs collection
- Kubernetes deployment
- production hardening beyond local development concerns
- moving the Go app into Docker Compose

## Architecture

The local stack will use the following data flow:

```text
Host Go app
  -> OTLP/gRPC traces
  -> OTLP/gRPC metrics
  -> OpenTelemetry Collector
      -> traces -> Tempo
      -> metrics -> Prometheus exporter endpoint
  -> Prometheus scrapes Collector metrics endpoint
  -> Grafana queries Prometheus and Tempo
```

The Go application will not expose `/metrics`. Instead, it will expose only business endpoints and push telemetry to the Collector. The Collector will convert metrics into a Prometheus-scrapable endpoint for Prometheus to pull.

## Components

### Go Demo Application

The new Go app will live alongside the existing demos and remain separate from `cmd/oteldemo` so the repository keeps both:

- a small OTel Prometheus-exporter demo
- a full OTLP local-stack demo

The application will:

- expose business endpoints such as `/work` and `/checkout`
- create at least one root HTTP span per request
- create a nested internal span so traces are visibly structured in Tempo
- record request count, request failures, and request duration metrics
- export both traces and metrics to the Collector using OTLP/gRPC
- shut down providers cleanly so batched telemetry is flushed

### OpenTelemetry Collector

The Collector will:

- receive OTLP/gRPC from the application
- batch telemetry
- export traces to Tempo over OTLP
- expose a Prometheus exporter endpoint for metrics

This keeps Prometheus in its natural pull model while allowing the Go application to stay OTLP-only.

### Prometheus

Prometheus will scrape only the Collector’s Prometheus exporter endpoint. It will not scrape the Go application directly.

### Tempo

Tempo will receive traces from the Collector and store them for local exploration through Grafana.

### Grafana

Grafana will be provisioned with:

- a Prometheus data source
- a Tempo data source

This allows a local user to inspect both metrics and traces without manual UI setup.

## File Structure

### New files

- `cmd/otelstackdemo/main.go`
- `internal/otelstackdemo/app.go`
- `internal/otelstackdemo/app_test.go`
- `internal/otelstackdemo/telemetry.go`
- `internal/otelstackdemo/telemetry_test.go`
- `deploy/otelcol/config.yaml`
- `deploy/prometheus/prometheus.yml`
- `deploy/tempo/tempo.yaml`
- `deploy/grafana/provisioning/datasources/datasources.yaml`
- `docker-compose.yml`
- `docs/otel-local-stack.md`

### Existing files to update

- `README.md`
- optionally `docs/openTelemetry-all.md` with a pointer to the runnable stack

## Error Handling

For this demo, telemetry initialization failure should fail fast at startup instead of silently degrading. This keeps the stack easier to understand and avoids a confusing state where the server runs but no telemetry appears downstream.

## Verification Strategy

Verification will be split into two layers:

- automated tests for app behavior and telemetry initialization helpers
- manual local stack verification using Docker Compose plus `go run`

The manual verification path will be:

1. `docker compose up -d`
2. `go run ./cmd/otelstackdemo`
3. send requests to the app
4. confirm Prometheus sees the exported metrics
5. confirm Tempo receives traces
6. confirm Grafana can query both data sources

## Trade-offs

### Why not expose `/metrics` from the app?

The user explicitly wants an OTLP-first example where the application does not expose metrics directly. Using the Collector’s Prometheus exporter preserves that goal while still integrating naturally with Prometheus.

### Why not run the app in Docker Compose?

Keeping the app on the host machine improves the local development loop and makes the OTLP wiring easier to understand. The architecture remains the same if the app is later moved into Compose.

### Why no logs in the first iteration?

Logs in OpenTelemetry Go remain less mature than traces and metrics. Adding logs now would increase complexity without improving the core goal of demonstrating a stable end-to-end metrics + traces pipeline.
