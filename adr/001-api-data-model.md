# ADR 001: API and Data-Model Architecture

**Status:** Accepted  
**Date:** 2026-10-26

## Context

The project requires a read-only HTTP API for a fictional port dataset. The key design
decisions involve the resource model, response format, error handling, and how the API
contract is documented and enforced.

## Decision

- **Standard library HTTP**: Use `net/http` with Go 1.22+ pattern-based routing
  (`ServeMux` path parameters). No third-party router is needed for this surface area.
- **Versioned API**: All routes are prefixed `/api/v1/` to allow non-breaking additions.
- **JSON responses**: All responses use `application/json`. Errors use
  `application/problem+json` (RFC 9457) with stable `type`, `title`, `status`, and
  `detail` fields. No implementation detail leaks in error bodies.
- **Pagination**: All collection endpoints accept `limit` (default 20, max 100) and
  `offset` (default 0). Responses include `total`, `limit`, and `offset` fields.
- **Deterministic ordering**: Collections are ordered by stable ID or time fields to
  ensure reproducible responses.
- **Filtering**: Per-resource query parameters are provided where useful (kind, flag
  country, facility, time bounds, commodity, etc.). Invalid filter values return 400.
- **OpenAPI 3.1.0**: `openapi/openapi.yaml` is the sole canonical contract. It is
  validated in CI via Vacuum. Insomnia exports and examples are aligned with it.
- **Embedded data**: The dataset is embedded in the server binary via Go `embed`. No
  database, random generator, or external fetch is involved. See ADR 004.

## Consequences

- The standard library mux is sufficient and avoids a dependency. Path value extraction
  (`r.PathValue`) requires Go 1.22+.
- A single OpenAPI source of truth prevents drift between documentation and
  implementation.
- Bounded pagination prevents unbounded response sizes.
- Deterministic ordering makes the API predictable and testable.
