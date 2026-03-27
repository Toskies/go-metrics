# OTel Local Stack Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a runnable local OpenTelemetry stack where a host-run Go app sends traces and metrics over OTLP to a Docker Compose stack containing Collector, Prometheus, Tempo, and Grafana.

**Architecture:** Add a new `otelstackdemo` Go app that emits OTLP traces and metrics only, with no application `/metrics` endpoint. Use the Collector as the central hub, exporting traces to Tempo and exposing a Prometheus scrape endpoint for metrics that Prometheus scrapes and Grafana visualizes.

**Tech Stack:** Go, OpenTelemetry Go SDK, OTLP gRPC exporters, OpenTelemetry Collector, Prometheus, Tempo, Grafana, Docker Compose, standard library `net/http` and `testing`

---

### File Structure

**Files:**
- Create: `cmd/otelstackdemo/main.go`
- Create: `internal/otelstackdemo/app.go`
- Create: `internal/otelstackdemo/app_test.go`
- Create: `internal/otelstackdemo/telemetry.go`
- Create: `internal/otelstackdemo/telemetry_test.go`
- Create: `deploy/otelcol/config.yaml`
- Create: `deploy/prometheus/prometheus.yml`
- Create: `deploy/tempo/tempo.yaml`
- Create: `deploy/grafana/provisioning/datasources/datasources.yaml`
- Create: `docker-compose.yml`
- Create: `docs/otel-local-stack.md`
- Modify: `README.md`
- Optionally modify: `docs/openTelemetry-all.md`

**Responsibilities:**
- `cmd/otelstackdemo/main.go`: process entrypoint and graceful shutdown wiring
- `internal/otelstackdemo/app.go`: HTTP handlers, spans, and metric recording
- `internal/otelstackdemo/app_test.go`: business endpoint behavior tests
- `internal/otelstackdemo/telemetry.go`: resource, exporters, providers, and setup/shutdown
- `internal/otelstackdemo/telemetry_test.go`: telemetry setup unit tests for configuration behavior
- `deploy/*`: local stack configuration
- `docker-compose.yml`: Compose orchestration for Collector, Prometheus, Tempo, and Grafana
- `docs/otel-local-stack.md`: operator instructions and validation steps

### Task 1: Add Failing Tests For The New App

**Files:**
- Create: `internal/otelstackdemo/app_test.go`
- Create: `internal/otelstackdemo/telemetry_test.go`

- [ ] **Step 1: Write the failing app behavior test**

Create a test that:
- builds the new app handler
- calls `/work`
- expects HTTP 200
- calls `/checkout?fail=1`
- expects HTTP 500

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/otelstackdemo -run 'TestAppRoutes|TestTelemetryConfig' -v`
Expected: FAIL because the package and functions do not exist yet

- [ ] **Step 3: Write the failing telemetry config test**

Create a test that verifies setup fails clearly when the OTLP endpoint is missing or malformed according to the chosen configuration helper API.

- [ ] **Step 4: Run test to verify it fails**

Run: `go test ./internal/otelstackdemo -run 'TestAppRoutes|TestTelemetryConfig' -v`
Expected: FAIL with missing symbols or missing implementation

### Task 2: Implement The OTLP Telemetry Layer

**Files:**
- Create: `internal/otelstackdemo/telemetry.go`
- Test: `internal/otelstackdemo/telemetry_test.go`

- [ ] **Step 1: Write minimal telemetry implementation**

Implement:
- resource configuration
- OTLP trace exporter setup
- OTLP metric exporter setup
- tracer provider and meter provider creation
- shutdown behavior

- [ ] **Step 2: Run targeted tests**

Run: `go test ./internal/otelstackdemo -run TestTelemetryConfig -v`
Expected: PASS

### Task 3: Implement The HTTP Demo Application

**Files:**
- Create: `internal/otelstackdemo/app.go`
- Create: `cmd/otelstackdemo/main.go`
- Test: `internal/otelstackdemo/app_test.go`

- [ ] **Step 1: Write minimal app implementation**

Implement:
- `/work` success endpoint
- `/checkout` endpoint with optional failure path
- root spans for requests
- a nested internal span
- request counter, failure counter, and duration histogram

- [ ] **Step 2: Run targeted tests**

Run: `go test ./internal/otelstackdemo -run TestAppRoutes -v`
Expected: PASS

- [ ] **Step 3: Run package tests**

Run: `go test ./internal/otelstackdemo -v`
Expected: PASS

### Task 4: Add The Local Observability Stack

**Files:**
- Create: `deploy/otelcol/config.yaml`
- Create: `deploy/prometheus/prometheus.yml`
- Create: `deploy/tempo/tempo.yaml`
- Create: `deploy/grafana/provisioning/datasources/datasources.yaml`
- Create: `docker-compose.yml`

- [ ] **Step 1: Write Collector configuration**

Configure:
- OTLP receiver
- batch processor
- OTLP exporter to Tempo
- Prometheus exporter endpoint for metrics

- [ ] **Step 2: Write Prometheus configuration**

Configure Prometheus to scrape only the Collector’s metrics endpoint.

- [ ] **Step 3: Write Tempo and Grafana provisioning**

Configure:
- Tempo OTLP receiver
- Grafana data sources for Prometheus and Tempo

- [ ] **Step 4: Run Compose config validation**

Run: `docker compose config`
Expected: valid merged configuration with no YAML errors

### Task 5: Add Run Documentation And Entry Links

**Files:**
- Create: `docs/otel-local-stack.md`
- Modify: `README.md`
- Optionally modify: `docs/openTelemetry-all.md`

- [ ] **Step 1: Write local run documentation**

Document:
- startup order
- environment variables
- sample requests
- Grafana, Prometheus, Tempo URLs
- what success looks like

- [ ] **Step 2: Add README entry**

Add a concise section linking to the new local-stack demo and doc.

### Task 6: Verify End To End

**Files:**
- Verify: all new files above

- [ ] **Step 1: Run Go tests**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 2: Validate Compose**

Run: `docker compose config`
Expected: PASS

- [ ] **Step 3: Start the local stack**

Run: `docker compose up -d`
Expected: all four services start successfully

- [ ] **Step 4: Run the Go app**

Run: `go run ./cmd/otelstackdemo`
Expected: app starts and connects to `localhost:4317`

- [ ] **Step 5: Send sample traffic**

Run example requests against `/work` and `/checkout`.
Expected: successful and failed requests generate traces and metrics

- [ ] **Step 6: Confirm observability outputs**

Verify:
- Prometheus target for Collector is up
- metrics appear in Prometheus queries
- traces appear in Tempo via Grafana Explore

- [ ] **Step 7: Review git status**

Run: `git status --short`
Expected: only intended changes for the new demo, docs, and deployment files
