# ADR 003: Observability and Operational Diagnostics

**Status:** Accepted  
**Date:** 2026-10-26

## Context

The project needs sufficient operational visibility for demonstration and debugging
without introducing a heavy observability stack (no Prometheus, Jaeger, etc.).

## Decision

- **Structured logging**: Use `log/slog` (Go standard library, Go 1.21+) with JSON
  output at INFO level by default. All request logs include method, path, status,
  duration, and remote address.
- **No sensitive data in logs**: Private keys, SVIDs, and credential material are never
  logged. Auth decisions log the SPIFFE ID and outcome but not certificate contents.
- **Startup/shutdown messages**: The server logs dataset sizes, configuration mode, and
  listen address at startup. Graceful shutdown is logged.
- **Health endpoints**: `GET /api/v1/health/live` (liveness) and
  `GET /api/v1/health/ready` (readiness) are unauthenticated and return minimal JSON.
  They do not expose infrastructure details, software versions, or internal state.
- **No external backend required**: Logs go to stdout in JSON format, suitable for
  container log aggregation. No sidecar, agent, or external service is required.

## Consequences

- slog is part of the Go standard library — no additional dependency.
- JSON structured logs integrate with standard aggregation pipelines (Loki, CloudWatch,
  etc.) without additional configuration.
- The initial scope does not include metrics or distributed tracing. Both can be added
  later via OpenTelemetry without changing the core application logic.
