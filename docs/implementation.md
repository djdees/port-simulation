# New Orleans Port Information Service — Implementation Plan

**Revision:** 2026-10-29  
**Status:** Approved — proceeding to implementation

This document records the implementation plan for the New Orleans Port Information Service,
reproducing the state described in [Prompt-2.md](../Prompt-2.md) (revision 2026-10-26) from a
clean workspace. Where this plan and Prompt-2.md differ, Prompt-2.md takes precedence.

---

## Background

The project is an engineering demonstration of:

- A read-only Go HTTP/JSON API server embedding a fictional port dataset
- A separate Go CLI client with periodic collector and local snapshot endpoint
- SPIFFE/SPIRE workload identity and mutual TLS
- Containerized deployment using Chainguard static images and multi-stage Docker builds

`make check` passed at the last verified state (2026-10-26, Go 1.27.3 in the original
environment). SPIRE integration is implemented but not integration-tested against a live
deployment. Docker image builds have a known Chainguard registry `403` issue in the original
environment.

> **Go version note:** The original prompt targets Go `1.27.3`. The machine running this
> implementation has Go `1.22.4`. The module is written to compile on `1.22` and is
> straightforward to upgrade when a newer toolchain is available.

---

## Open Questions / Resolved Decisions

| Question | Decision |
|---|---|
| Go version | Target `go 1.22` (installed); straightforward to bump when 1.27.x is available |
| `docs/` casing | Lowercase `docs/` on Linux (Prompt-2 `Docs/` was a macOS worktree artifact) |
| Module path | `github.com/djdees/port-simulation` |
| `Implementation.md` | Not reproduced separately; `docs/implementation.md` serves that purpose |
| `spire-setup.md` | Produced as optional local lab guide; no automated SPIRE deployment |

---

## Repository Layout

```text
port-simulation/
├── cmd/
│   ├── server/
│   │   └── main.go
│   └── client/
│       └── main.go
├── internal/
│   ├── client/
│   │   ├── client.go       (SPIFFE transport, mTLS)
│   │   ├── query.go        (one-shot subcommands)
│   │   ├── collector.go    (periodic CSV collector)
│   │   ├── snapshot.go     (local GET /api/v1/latest/{query})
│   │   ├── config/
│   │   │   └── config.go
│   │   └── client_test.go
│   ├── domain/
│   │   ├── models.go
│   │   ├── validate.go
│   │   └── validate_test.go
│   └── server/
│       ├── auth/
│       │   └── auth.go     (SPIFFE source, mTLS, policy)
│       ├── config/
│       │   └── config.go
│       ├── data/
│       │   ├── port.json
│       │   └── embed.go
│       └── handler/
│           ├── handler.go  (mux, middleware)
│           ├── vessels.go
│           ├── barges.go
│           ├── tows.go
│           ├── cargo.go
│           ├── facilities.go
│           ├── terminals.go
│           ├── berths.go
│           ├── events.go
│           ├── health.go
│           └── handler_test.go
├── openapi/
│   └── openapi.yaml
├── adr/
│   ├── 001-api-data-model.md
│   ├── 002-spiffe-trust-model.md
│   ├── 003-observability.md
│   ├── 004-dataset-storage.md
│   └── 005-container-deployment.md
├── docs/
│   ├── implementation.md   (this file)
│   └── spire-integration.md
├── examples/
│   └── examples.md
├── insomnia/
│   └── port-service.json
├── docker/
│   ├── Dockerfile.server
│   ├── Dockerfile.client
│   └── docker-compose.yml
├── spire-setup.md
├── .gitignore
├── Makefile
├── README.md
├── go.mod
└── go.sum
```

---

## Component Plan

### 1. Go Module and Root Files

**`go.mod`**

- Module: `github.com/djdees/port-simulation`
- Go: `1.22`
- Runtime dependency: `github.com/spiffe/go-spiffe/v2 v2.9.0`
- Indirect dependencies locked via `go.sum`

**`.gitignore`**

Excludes binaries, build artifacts, test output and coverage files, local env/config files,
secrets, IDE directories. `go.sum` is *included* (not excluded).

**`Makefile`**

| Target | Description |
|---|---|
| `help` | List all targets |
| `fmt` | `gofmt` / `goimports` check |
| `vet` | `go vet ./...` |
| `lint` | `golangci-lint run` |
| `test` | `go test ./...` |
| `test-race` | `go test -race ./...` |
| `build` | Build server + client |
| `build-server` | Build server binary only |
| `build-client` | Build client binary only |
| `run-server` | Run server locally (requires env) |
| `run-client` | Run client locally (requires env) |
| `openapi-check` | Validate `openapi/openapi.yaml` with Vacuum |
| `security` | `govulncheck` + `gosec` |
| `check` | `fmt vet lint test test-race openapi-check security build` |
| `clean` | Remove binaries and artifacts |
| `docker-build-server` | Build server Docker image |
| `docker-build-client` | Build client Docker image |
| `compose-up` | Start server via Compose |
| `compose-client` | Run one-shot client via Compose |
| `install-tools` | Install pinned external tools |

Pinned tool versions: `golangci-lint v2.15.0`, `vacuum v0.35.4`, `govulncheck v1.1.4`,
`gosec v2.30.0`.

---

### 2. Domain Models

**`internal/domain/models.go`**

Core types and enums:

- `Vessel` — ocean freighters, tankers, offshore service vessels; fields: id, name, IMO,
  flag country, kind, MMSI, call sign, owner
- `Barge` — river barges; fields: id, name, owner, capacity, capability
- `Tow` — barge fleet; fields: id, towboat vessel ID, barge IDs
- `Cargo` — commodity records; fields: id, commodity, origin, destination, carrier vessel or
  barge ID, weight, handling notes
- `Facility` — port facility; fields: id, name, kind (commercial/industrial/service)
- `Terminal` — terminal within a facility; fields: id, name, facility ID
- `Berth` — berth within a terminal; fields: id, name, terminal ID, max draft, max length
- `TrafficEvent` — port traffic events; fields: id, event type, subject vessel/barge ID,
  facility ID, berth ID, cargo IDs, time (UTC RFC3339), notes

Enums: `VesselKind`, `BargeCapability`, `FacilityKind`, `EventType`, `CommodityType`

**`internal/domain/validate.go`**

Validates full dataset at startup and in tests:

- Required fields and stable unique IDs
- Enum membership
- Numeric constraints (capacity ≥ 0, draft > 0, etc.)
- Vessel↔tow and barge↔tow cross-references
- Cargo carrier references point to known vessels or barges
- Facility/terminal/berth hierarchy integrity
- Event subject, facility, berth, and cargo cross-references

---

### 3. Embedded Dataset (`internal/server/data/port.json`)

Small, deterministic, fictional dataset. Stable IDs use prefix conventions:

| Prefix | Type |
|---|---|
| `VSL-xxx` | Vessel |
| `BGE-xxx` | Barge |
| `TOW-xxx` | Tow |
| `CGO-xxx` | Cargo |
| `FAC-xxx` | Facility |
| `TRM-xxx` | Terminal |
| `BTH-xxx` | Berth |
| `EVT-xxx` | Traffic Event |

Approximate record counts: 6 vessels, 4 barges, 2 tows, 8 cargo records, 3 facilities,
6 terminals, 10 berths, 10 traffic events. All data is fictional and labeled as simulated.

Loaded via `//go:embed port.json`, parsed and validated at server startup by `embed.go`.

---

### 4. Server

**Configuration** (`internal/server/config/config.go`)

| Variable | Flag | Default | Description |
|---|---|---|---|
| `PORT_SERVER_LISTEN` | `--listen` | `127.0.0.1:8443` | Listen address |
| `PORT_SPIFFE_ENABLED` | `--spiffe` | `true` | Enable SPIFFE/mTLS |
| `SPIFFE_ENDPOINT_SOCKET` | — | — | Workload API socket |
| `PORT_AUTH_POLICY` | `--policy` | — | Path to authorization policy JSON |

**Authorization** (`internal/server/auth/auth.go`)

- SPIFFE Workload API X.509 source (kept alive for rotation)
- mTLS TLS config restricted to configured client SPIFFE IDs
- JSON policy: exact client SPIFFE ID → permissions (`port:read`)
- SPIFFE-off mode: loopback-only, HTTP only, logs warning, no remote access

**HTTP handlers** (`internal/server/handler/`)

- Middleware: structured request logging, timeout, max request size, problem+JSON error shaping
- All collection endpoints: `limit` (default 20, max 100) + `offset` pagination, deterministic ordering
- `application/problem+json` (RFC 9457) error responses with stable error codes
- No implementation detail leakage in error bodies

**Filters per endpoint:**

| Endpoint | Filters |
|---|---|
| `GET /api/v1/vessels` | `kind`, `flag_country` |
| `GET /api/v1/arrivals` | `facility_id`, `since`, `until` |
| `GET /api/v1/departures` | `facility_id`, `since`, `until` |
| `GET /api/v1/barges` | `capability`, `tow_id` |
| `GET /api/v1/tows` | `towboat_id`, `barge_id` |
| `GET /api/v1/cargo` | `commodity`, `origin`, `destination`, `vessel_id`, `barge_id` |
| `GET /api/v1/facilities` | `kind` |
| `GET /api/v1/terminals` | `facility_id` |
| `GET /api/v1/berths` | `terminal_id` |

---

### 5. Client

**Configuration** (`internal/client/config/config.go`)

| Variable | Flag | Description |
|---|---|---|
| `PORT_SERVER_URL` | `--server` | Server base URL |
| `PORT_SPIFFE_ENABLED` | `--spiffe` | Enable SPIFFE/mTLS |
| `SPIFFE_ENDPOINT_SOCKET` | — | Workload API socket |
| `PORT_SERVER_SPIFFE_ID` | `--server-id` | Expected server SPIFFE ID |
| `PORT_LOCAL_READER_SPIFFE_IDS` | — | Allowed readers for snapshot endpoint |
| `PORT_COLLECTOR_INTERVAL` | — | Tick interval (default 30s) |
| `PORT_COLLECTOR_OUTPUT_DIR` | — | CSV output directory |

**One-shot subcommands:** `query`, `list-arrivals`, `get-vessel <id>`, `filter-cargo`

**Collector** — five presets: `arrivals`, `departures`, `vessels`, `cargo`, `berths`.
One preset per 30-second tick; each preset repeats every 150 seconds. Collection starts
immediately. Successful rows appended to per-preset CSV with UTC `collected_at` column.
Failed requests logged; no partial rows written.

**Snapshot endpoint** — `GET /api/v1/latest/{query}` served by the client process on loopback only.
Returns JSON: `{"query_id":"...", "collected_at":"...", "items":[...]}`.
Returns `204 No Content` before first successful snapshot. No POST refresh. Requires mTLS +
`PORT_LOCAL_READER_SPIFFE_IDS` when SPIFFE is enabled.

---

### 6. OpenAPI Specification (`openapi/openapi.yaml`)

- Version: OpenAPI 3.1.0
- All routes from §5 of Prompt-2.md documented
- `mutualTLS` security scheme declared with explanatory description
- `application/problem+json` error schema on all 4xx/5xx responses
- Representative examples for each endpoint
- Validated with Vacuum v0.35.4 (`make openapi-check`)

---

### 7. Docker

**`docker/Dockerfile.server`** and **`docker/Dockerfile.client`**

- Multi-stage: `golang:1.22` builder → Chainguard `cgr.dev/chainguard/static` runtime
- Non-root UID/GID 65532
- Pinned runtime image digest
- No secrets or SVIDs in image layers

**`docker/docker-compose.yml`**

- Starts server by default (`make compose-up`)
- Client is a one-shot explicit invocation (`make compose-client ARGS="query arrivals"`)
- Requires external env: `SPIFFE_SOCKET`, `PORT_AUTH_POLICY_FILE`, `PORT_SERVER_SPIFFE_ID`
- Does **not** deploy SPIRE

---

### 8. Documentation

| File | Contents |
|---|---|
| `README.md` | Full introduction: purpose, architecture, layout, build, config, Docker, SPIFFE, testing |
| `adr/001-api-data-model.md` | API and data-model architecture decision |
| `adr/002-spiffe-trust-model.md` | SPIFFE/SPIRE integration and workload trust model |
| `adr/003-observability.md` | Logging, health checks, observability scope |
| `adr/004-dataset-storage.md` | Embedded JSON dataset decision |
| `adr/005-container-deployment.md` | Container architecture and external SPIRE dependency |
| `docs/spire-integration.md` | External SPIRE requirements, workload registrations, selectors, rotation |
| `spire-setup.md` | Optional local SPIRE lab (SPIRE v1.15.5, join-token attestation, Docker labels) |
| `examples/examples.md` | Representative curl/client command examples |
| `insomnia/port-service.json` | Importable Insomnia collection (no credentials) |

---

## Verification Plan

### Automated (no SPIRE required)

```sh
go version
make install-tools
make check
```

`make check` runs: `fmt`, `vet`, `lint`, `test`, `test-race`, `openapi-check`, `security`,
`build-server`, `build-client`.

### Compose Config Validation

```sh
SPIFFE_SOCKET=/tmp/spire-agent.sock \
PORT_AUTH_POLICY_FILE=/absolute/path/to/authorization-policy.json \
PORT_SERVER_SPIFFE_ID=spiffe://port.example.test/ns/demo/sa/server \
docker compose -f docker/docker-compose.yml config --quiet
```

### Loopback Smoke Test (SPIFFE disabled, development only)

```sh
PORT_SPIFFE_ENABLED=false ./bin/server --listen 127.0.0.1:8080
PORT_SPIFFE_ENABLED=false ./bin/client query arrivals --server http://127.0.0.1:8080
```

### Not Verified (requires real SPIRE)

- Workload API socket access and SVID issuance
- mTLS handshake with real X.509-SVIDs
- Authorization against actual SPIRE-issued identities
- SVID/bundle rotation behavior
- Docker image builds (pending Chainguard registry access)

---

## Suggested Commit Boundaries

```
feat(api): add fictional port information service
feat(client): add query collector and CSV snapshots
feat(security): add SPIFFE workload identity authorization
build(docker): add Chainguard server and client images
docs: document SPIRE lab setup and API examples
```

No commits are authorized by this plan. The repository owner performs all staging and commits.
