# New Orleans Port Information Service

> **SIMULATED DATA ONLY.** All vessels, barges, cargo, facilities, terminals, berths,
> events, identifiers, and schedules in this project are entirely fictional. This is an
> engineering demonstration. It has no affiliation with or endorsement by the Port of
> New Orleans or any actual port operator.

An engineering demonstration of:
- A read-only Go HTTP/JSON API server with an embedded fictional port dataset
- A separate Go CLI client with a periodic collector and local snapshot endpoint
- **SPIFFE/SPIRE** workload identity and mutual TLS (mTLS)
- Containerized deployment using Chainguard static images

## Architecture

```
┌───────────────────────────────────────────────────────────┐
│                   port-simulation                         │
│                                                           │
│   ┌─────────────────┐          ┌─────────────────────┐   │
│   │   cmd/server    │  mTLS    │    cmd/client        │   │
│   │   (port API)    │◄────────►│  (CLI + collector)  │   │
│   │                 │          │                       │   │
│   │  GET /api/v1/*  │          │  query | collect      │   │
│   │  health/live    │          │  GET /latest/{query}  │   │
│   │  health/ready   │          │  (loopback only)      │   │
│   └────────┬────────┘          └──────────┬────────────┘  │
│            │                              │               │
│   ┌────────▼──────────────────────────────▼────────┐     │
│   │          SPIRE Agent (external)                 │     │
│   │       Workload API Unix socket                  │     │
│   │       X.509-SVID issuance + rotation            │     │
│   └─────────────────────────────────────────────────┘     │
└───────────────────────────────────────────────────────────┘
```

The server embeds the fictional dataset in its binary. No database, external data
source, or runtime data mount is required.

## Repository Layout

```
port-simulation/
├── cmd/server/         Server entrypoint
├── cmd/client/         Client entrypoint
├── internal/
│   ├── domain/         Data models and dataset validation
│   ├── server/
│   │   ├── auth/       SPIFFE auth, mTLS, authorization policy
│   │   ├── config/     Server configuration
│   │   ├── data/       Embedded port.json + loader
│   │   └── handler/    HTTP handlers and tests
│   └── client/
│       ├── config/     Client configuration
│       ├── client.go   HTTP transport (SPIFFE or plain)
│       ├── query.go    One-shot query commands
│       ├── collector.go Periodic CSV collector
│       └── snapshot.go Local snapshot endpoint
├── openapi/            openapi.yaml (canonical API contract)
├── adr/                Architectural decision records
├── docs/               Supporting documentation
├── examples/           API usage examples
├── insomnia/           Importable Insomnia collection
├── docker/             Dockerfiles and Compose
├── spire-setup.md      Optional local SPIRE lab guide
└── Makefile
```

## Prerequisites

- **Go 1.24+** (go-spiffe v2.9.0 requires Go 1.24; install from https://go.dev/dl/)
- **Docker** (for image builds and Compose)
- **make**

External development tools (installed via `make install-tools`):
- golangci-lint v2.15.0
- vacuum v0.35.4
- govulncheck v1.1.4
- gosec v2.30.0

## Quick Start (SPIFFE-off, loopback only)

> This mode is for local development only. SPIFFE/mTLS is disabled. The server and
> client are restricted to loopback addresses.

```sh
# Terminal 1: start server
make run-server

# Terminal 2: run client queries
make run-client ARGS="query arrivals"
make run-client ARGS="get-vessel VSL-001"
make run-client ARGS="filter-cargo commodity=petroleum"
```

## Build

```sh
make build          # builds both binaries → bin/server, bin/client
make build-server   # server only
make build-client   # client only
```

## Configuration

### Server

| Variable | Flag | Default | Description |
|---|---|---|---|
| `PORT_SERVER_LISTEN` | `--listen` | `127.0.0.1:8443` | Listen address |
| `PORT_SPIFFE_ENABLED` | `--spiffe` | `true` | Enable SPIFFE/mTLS |
| `SPIFFE_ENDPOINT_SOCKET` | — | (required when SPIFFE=true) | Workload API socket |
| `PORT_AUTH_POLICY` | `--policy` | (required when SPIFFE=true) | Auth policy JSON path |

### Client

| Variable | Flag | Description |
|---|---|---|
| `PORT_SERVER_URL` | `--server` | Server base URL (required) |
| `PORT_SPIFFE_ENABLED` | `--spiffe` | Enable SPIFFE/mTLS (default: true) |
| `SPIFFE_ENDPOINT_SOCKET` | — | Workload API socket |
| `PORT_SERVER_SPIFFE_ID` | `--server-id` | Expected server SPIFFE ID |
| `PORT_COLLECTOR_OUTPUT_DIR` | — | CSV output directory (default: `collector-output/`) |
| `PORT_COLLECTOR_INTERVAL` | — | Collection tick interval (default: `30s`) |
| `PORT_SNAPSHOT_LISTEN` | — | Snapshot endpoint listen (default: `127.0.0.1:8081`) |
| `PORT_LOCAL_READER_SPIFFE_IDS` | — | Comma-separated reader SPIFFE IDs |

## API

See [`openapi/openapi.yaml`](openapi/openapi.yaml) for the full contract.

| Endpoint | Description |
|---|---|
| `GET /api/v1/vessels` | List vessels (filter: `kind`, `flag_country`) |
| `GET /api/v1/vessels/{id}` | Get vessel by ID |
| `GET /api/v1/arrivals` | List arrivals (filter: `facility_id`, `since`, `until`) |
| `GET /api/v1/departures` | List departures |
| `GET /api/v1/barges` | List barges (filter: `capability`, `tow_id`) |
| `GET /api/v1/barges/{id}` | Get barge by ID |
| `GET /api/v1/tows` | List tows (filter: `towboat_id`, `barge_id`) |
| `GET /api/v1/tows/{id}` | Get tow by ID |
| `GET /api/v1/cargo` | List cargo (filter: `commodity`, `origin`, `destination`, `vessel_id`, `barge_id`) |
| `GET /api/v1/cargo/{id}` | Get cargo by ID |
| `GET /api/v1/facilities` | List facilities (filter: `kind`) |
| `GET /api/v1/facilities/{id}` | Get facility by ID |
| `GET /api/v1/terminals` | List terminals (filter: `facility_id`) |
| `GET /api/v1/terminals/{id}` | Get terminal by ID |
| `GET /api/v1/berths` | List berths (filter: `terminal_id`) |
| `GET /api/v1/berths/{id}` | Get berth by ID |
| `GET /api/v1/health/live` | Liveness (unauthenticated) |
| `GET /api/v1/health/ready` | Readiness (unauthenticated) |

Collections accept `limit` (default 20, max 100) and `offset` pagination.
Errors use `application/problem+json` (RFC 9457).

## Makefile Targets

```sh
make help           # List all targets
make fmt            # Check formatting
make vet            # go vet
make lint           # golangci-lint
make test           # Run tests (no SPIRE required)
make test-race      # Run tests with -race
make openapi-check  # Validate openapi.yaml with Vacuum
make security       # govulncheck + gosec
make check          # All checks + build
make clean          # Remove binaries and collector output
make install-tools  # Install pinned external tools
```

## Docker and Compose

```sh
# Build images (requires Chainguard registry access)
make docker-build-server
make docker-build-client

# Start server via Compose (requires SPIFFE_SOCKET, PORT_AUTH_POLICY_FILE, PORT_SERVER_SPIFFE_ID)
make compose-up

# Run one-shot client
make compose-client ARGS="query arrivals"

# Validate Compose config without starting services
SPIFFE_SOCKET=/tmp/spire-agent.sock \
PORT_AUTH_POLICY_FILE=/path/to/policy.json \
PORT_SERVER_SPIFFE_ID=spiffe://port.example.test/ns/demo/sa/server \
docker compose -f docker/docker-compose.yml config --quiet
```

## SPIFFE/SPIRE

The server and client authenticate using SPIFFE X.509-SVIDs issued by an external
SPIRE deployment. See [`docs/spire-integration.md`](docs/spire-integration.md) for:

- Trust domain and expected SPIFFE IDs
- Workload registration examples
- Authorization policy format
- Credential rotation behavior
- Integration test prerequisites

For a disposable local SPIRE lab, see [`spire-setup.md`](spire-setup.md).
**The main Compose file does not deploy SPIRE.**

## Testing

```sh
make test           # Unit and API tests — no SPIRE required
make test-race      # Same with race detector
```

SPIRE integration tests (Workload API access, real mTLS, SVID rotation) require a
live SPIRE deployment and are not part of `make test`.

## Security Checks

```sh
make security       # govulncheck + gosec
```

## Dataset

The dataset is in [`internal/server/data/port.json`](internal/server/data/port.json).
To update it:

1. Edit `port.json`
2. Run `make build-server` — validation runs at startup
3. Run `make test` — dataset integrity tests also run

All IDs use stable prefixes: `VSL-`, `BGE-`, `TOW-`, `CGO-`, `FAC-`, `TRM-`, `BTH-`, `EVT-`.

## Troubleshooting

| Symptom | Likely cause |
|---|---|
| `SPIFFE_ENDPOINT_SOCKET is required` | SPIFFE mode is on; set the env var or use `--spiffe=false` |
| `loopback listen address required` | SPIFFE-off mode does not allow non-loopback; use `--spiffe=true` or change the address |
| `403 Forbidden` on API request | Client SPIFFE ID not in authorization policy |
| `server returned 400` | Invalid query parameter (check enum values) |
| Docker build `403 Forbidden` | Chainguard registry auth issue; check registry credentials |

## Verification State (as of 2026-11-12)

- `go build ./...` — ✅ passes
- `go test ./...` — ✅ 19 tests pass
- `go test -race ./...` — ✅ passes
- `go vet ./...` — ✅ passes
- OpenAPI validation — requires `make install-tools` then `make openapi-check`
- govulncheck / gosec — requires `make install-tools` then `make security`
- Docker image builds — not verified (pending Chainguard registry access)
- SPIRE integration — not verified (requires live SPIRE environment)
