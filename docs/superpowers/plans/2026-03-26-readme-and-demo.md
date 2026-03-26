# README Summary And Go Demo Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Add a concise repository README plus a runnable Go demo that shows the difference between Prometheus `client_golang` and OpenTelemetry metrics exposure.

**Architecture:** Keep the repository small and didactic. Put the long-form explanation in `docs/`, add a concise `README.md` that links to it, and implement two tiny HTTP servers under `cmd/` with focused packages under `internal/` so we can test the instrumentation behavior directly.

**Tech Stack:** Go, `github.com/prometheus/client_golang`, `go.opentelemetry.io/otel`, `go.opentelemetry.io/otel/sdk/metric`, `go.opentelemetry.io/otel/exporters/prometheus`, standard library `net/http` and `testing`

---

### File Structure

**Files:**
- Create: `README.md`
- Create: `go.mod`
- Create: `go.sum`
- Create: `cmd/promdemo/main.go`
- Create: `cmd/oteldemo/main.go`
- Create: `internal/promdemo/server.go`
- Create: `internal/promdemo/server_test.go`
- Create: `internal/oteldemo/server.go`
- Create: `internal/oteldemo/server_test.go`
- Keep: `docs/go-metrics-prometheus-opentelemetry.md`

**Responsibilities:**
- `README.md`: repository landing page with a short comparison, links, and run commands
- `cmd/promdemo/main.go`: start a tiny HTTP server exposing Prometheus `client_golang` metrics
- `cmd/oteldemo/main.go`: start a tiny HTTP server exposing OpenTelemetry metrics via the Prometheus exporter
- `internal/promdemo/server.go`: reusable Prometheus demo handler setup
- `internal/promdemo/server_test.go`: verify the Prometheus demo emits expected metrics after sample traffic
- `internal/oteldemo/server.go`: reusable OpenTelemetry demo handler setup
- `internal/oteldemo/server_test.go`: verify the OpenTelemetry demo emits expected metrics after sample traffic

### Task 1: Add Repository README Summary

**Files:**
- Create: `README.md`
- Reference: `docs/go-metrics-prometheus-opentelemetry.md`

- [ ] **Step 1: Write the README content**

Include:
- repository purpose
- short comparison table for `client_golang` vs OpenTelemetry
- link to the long-form document
- commands to run both demos

- [ ] **Step 2: Review the README for alignment**

Check that the README is shorter than the long-form document and clearly points readers to the deeper explanation.

### Task 2: Build The Prometheus Demo With TDD

**Files:**
- Create: `go.mod`
- Create: `internal/promdemo/server.go`
- Create: `internal/promdemo/server_test.go`
- Create: `cmd/promdemo/main.go`

- [ ] **Step 1: Write the failing test**

Create `internal/promdemo/server_test.go` with a test that:
- creates the demo handler
- sends one request to `/work`
- fetches `/metrics`
- expects `demo_http_requests_total` and `demo_http_request_duration_seconds` in the response body

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/promdemo -run TestHandlerExposesMetrics -v`
Expected: FAIL because the package or handler does not exist yet

- [ ] **Step 3: Write minimal implementation**

Create a small handler package that:
- uses a dedicated Prometheus registry
- registers request counter and latency histogram
- serves `/work` and `/metrics`

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/promdemo -run TestHandlerExposesMetrics -v`
Expected: PASS

- [ ] **Step 5: Add the runnable entry point**

Create `cmd/promdemo/main.go` to start the server on `:2112`.

### Task 3: Build The OpenTelemetry Demo With TDD

**Files:**
- Create: `internal/oteldemo/server.go`
- Create: `internal/oteldemo/server_test.go`
- Create: `cmd/oteldemo/main.go`

- [ ] **Step 1: Write the failing test**

Create `internal/oteldemo/server_test.go` with a test that:
- creates the demo handler
- sends one request to `/work`
- fetches `/metrics`
- expects `demo_http_requests_total` and histogram output in the response body

- [ ] **Step 2: Run test to verify it fails**

Run: `go test ./internal/oteldemo -run TestHandlerExposesMetrics -v`
Expected: FAIL because the package or handler does not exist yet

- [ ] **Step 3: Write minimal implementation**

Create a small OpenTelemetry handler package that:
- configures a Prometheus exporter and meter provider
- records request count and latency
- serves `/work` and `/metrics`

- [ ] **Step 4: Run test to verify it passes**

Run: `go test ./internal/oteldemo -run TestHandlerExposesMetrics -v`
Expected: PASS

- [ ] **Step 5: Add the runnable entry point**

Create `cmd/oteldemo/main.go` to start the server on `:2113`.

### Task 4: Full Verification

**Files:**
- Verify: `README.md`
- Verify: `cmd/promdemo/main.go`
- Verify: `cmd/oteldemo/main.go`
- Verify: `internal/promdemo/server_test.go`
- Verify: `internal/oteldemo/server_test.go`

- [ ] **Step 1: Run focused tests**

Run: `go test ./internal/promdemo ./internal/oteldemo -v`
Expected: PASS

- [ ] **Step 2: Run full test suite**

Run: `go test ./...`
Expected: PASS

- [ ] **Step 3: Review git status**

Run: `git status --short`
Expected: only intended new files for docs, README, and Go demo
