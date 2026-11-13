# Project Prompt 2: New Orleans Port Information Service

**Revision date:** 2026-10-26  
**Purpose:** Current project brief and reproducibility handoff. This document carries forward the original requirements while recording decisions made and work completed to date.

Read [Prompt.md](Prompt.md) for the original full brief and [Implementation.md](Implementation.md) for the staged plan and verification log. Where requirements differ, the decisions and scope changes in this document supersede the original prompt. This is an updated brief, not authorization to weaken any security boundary or to create commits.

## 1. Objective and Scope

Maintain two independent Go applications for a fictionalized New Orleans-area port:

- A read-only server with a versioned JSON API over a deterministic, validated, fictional dataset embedded in its binary.
- An independent CLI and collector that can make one-shot requests, periodically collect API data into CSV, and expose its latest collected snapshot through a local HTTP GET endpoint.

The project is an engineering demonstration of API design, Go practices, containers, and SPIFFE/SPIRE identity. It is not a production port-management system and must not imply affiliation with or endorsement by the Port of New Orleans or any port operator.

All facilities, vessels, identifiers, cargo, events, schedules, and data are fictional. Do not add live tracking feeds, operational records, customer information, or real manifests. Do not add a database or CI/CD in this iteration.

### Approved Scope Change: Local SPIRE Lab Guide

The initial architecture keeps production SPIRE infrastructure externally managed. An optional developer-only procedure to create a disposable local SPIRE Server and Agent is now in scope as documentation in [spire-setup.md](spire-setup.md). This local lab is separate from the application Compose file, must not be treated as production guidance, and must not introduce an automatic insecure fallback. No SPIRE deployment has been verified by running that guide.

## 2. Decisions and Standards

These are owner-approved decisions:

- **Go:** target Go `1.27.3` as current on 2026-10-26; keep the module and Docker builder versions explicit and straightforward to upgrade.
- **Dependencies:** prefer the standard library. Use `github.com/spiffe/go-spiffe/v2` `v2.9.0` for Workload API, X.509-SVIDs, trust bundles, mTLS, and credential rotation.
- **HTTP:** use the standard library. The canonical server contract is OpenAPI 3.1.0 at `openapi/openapi.yaml`.
- **API:** keep the port server read-only. Use stable IDs, UTC RFC3339 timestamps, consistent problem responses, bounded pagination, filters, deterministic ordering, and minimal health responses.
- **SPIFFE mode:** enabled by default. `--spiffe=true|false` takes precedence over `PORT_SPIFFE_ENABLED=true|false`; if neither is supplied, SPIFFE is on.
- **SPIFFE-off mode:** an explicit development-only option is available before SPIRE is configured. It logs a warning and is restricted to loopback HTTP. There is no automatic fallback. It cannot be used for remote access or Compose deployment.
- **Authorization:** authentication and authorization are separate. The server requires exact configured client SPIFFE IDs and maps verified IDs to explicit permissions. The initial permission is `port:read`. The collector's reader allowlist is separate.
- **Tests:** ordinary unit/API tests run without SPIRE. Workload API, operational mTLS, SVID/bundle rotation, and end-to-end SPIRE checks are opt-in and require a real SPIRE environment.
- **Containers:** separate multi-stage server/client images with the Chainguard `static` runtime, non-root UID/GID 65532, and pinned runtime digest. Docker is the current runtime; Colima is a later evaluation.
- **Compose:** the project Compose file does not start SPIRE. It can start the server independently and run a one-shot client explicitly. It requires an externally available Workload API socket and valid policy in SPIFFE mode.
- **Repository:** preserve the existing `Docs/` location/casing as represented by this macOS worktree. The repository owner owns staging and commits. Do not commit, stage, amend, rebase, push, or publish unless separately requested.
- **Public project policy:** do not add a license or contribution policy without direction.

## 3. Current Implementation

The project currently contains:

- `cmd/server` and `cmd/client` executables.
- `internal/domain` models and cross-reference/enum validation.
- `internal/server/data/port.json`, embedded and validated by the server.
- A standard-library HTTP API, structured logs, timeouts, graceful shutdown, health routes, filtering, pagination, and safe problem responses.
- SPIFFE Workload API X.509 source integration, mTLS, exact identity checks, server/client authorization policy, explicit mode selection, and HTTPS enforcement for SPIFFE client requests.
- A one-shot client (`query`, `list-arrivals`, `get-vessel`, `filter-cargo`) and a sequential collector.
- Five collector presets, in order: `arrivals`, `departures`, `vessels`, `cargo`, `berths`.
- Collector cadence: one preset per 30-second tick; each preset repeats every 150 seconds. Collection begins immediately. Successful rows are appended to a per-preset CSV with UTC `collected_at`; failed requests are logged and do not create rows.
- Collector `GET /api/v1/latest/{query}`: returns JSON containing `query_id`, `collected_at`, and `items`; returns `204 No Content` before a successful snapshot; does not fetch or mutate state. There is no POST refresh endpoint. The collector listener is loopback-only and requires mTLS plus `PORT_LOCAL_READER_SPIFFE_IDS` when SPIFFE is enabled.
- `openapi/openapi.yaml`, safe Insomnia export, `examples/examples.md`, five ADRs, README, external SPIRE integration notes, and the optional local SPIRE lab guide.
- Makefile quality and build targets, `.gitignore`, separate Dockerfiles, and Compose configuration.

The SPIFFE-disabled mode is a testing/development aid only. It is not an identity implementation and must not become a normal deployment mode.

## 4. Domain and Dataset Requirements

Retain these linked concepts: commercial vessels, towboats, offshore-service vessels, barges, tows, cargo, facilities, terminals, berths, and traffic events. Keep the initial dataset small and deterministic.

The server must embed version-controlled JSON using Go `embed` and validate at startup and in tests:

- Required fields and stable unique IDs.
- Resource enums and numeric constraints.
- Vessel/towboat and barge/tow relationships.
- Cargo carrier references.
- Facility/terminal/berth relationships.
- Event subject, facility, berth, and cargo cross-references.

Keep all data fictional. No database, random generator, external data API, or manual data refresh is part of scope. Document dataset edits and rebuild steps.

## 5. API Contract

`openapi/openapi.yaml` is the sole canonical server API contract. Implement only useful read operations; keep implementation, Insomnia, examples, and OpenAPI aligned.

Current resources/routes:

- `GET /api/v1/vessels`, `GET /api/v1/vessels/{id}`
- `GET /api/v1/arrivals`, `GET /api/v1/departures`
- `GET /api/v1/barges`, `GET /api/v1/barges/{id}`
- `GET /api/v1/tows`, `GET /api/v1/tows/{id}`
- `GET /api/v1/cargo`, `GET /api/v1/cargo/{id}`
- `GET /api/v1/facilities`, `GET /api/v1/facilities/{id}`
- `GET /api/v1/terminals`, `GET /api/v1/terminals/{id}`
- `GET /api/v1/berths`, `GET /api/v1/berths/{id}`
- `GET /api/v1/health/live`, `GET /api/v1/health/ready`

Collections use `limit` (default 20, maximum 100) and `offset` (default 0), with deterministic ordering. Filters currently include vessel kind/flag country; event time bounds/facility; barge capability/tow; towboat/barge; cargo commodity/origin/destination/vessel/barge; facility kind; terminal facility; and berth terminal.

Use `application/problem+json` with stable error codes for HTTP errors. Do not leak implementation details. Keep health responses minimal and distinguish liveness from readiness. OpenAPI declares `mutualTLS` for protected endpoints and explains that SPIFFE identity authorization is separate from the OpenAPI contract.

The client's local `GET /api/v1/latest/{query}` is served by the client process, not the port server. Keep that distinction documented; it is not part of the port server's read-only API implementation.

## 6. SPIFFE/SPIRE Design

Normal and remote operation uses X.509-SVIDs and trust bundles obtained only from the SPIFFE Workload API. Use the maintained go-spiffe source and mTLS helpers and keep the source alive for listeners/transports to support rotation. Fail closed for missing identity, trust bundle, expected IDs, or authorization policy. Never trust identity in HTTP headers/query/body, use `InsecureSkipVerify`, create a general bypass, or silently fall back to plaintext.

The server authorization JSON maps exact client SPIFFE IDs to allowed operations, initially `port:read`. The TLS layer also limits accepted client IDs. The client verifies the configured server SPIFFE ID. The collector reader allowlist is `PORT_LOCAL_READER_SPIFFE_IDS` and is distinct from the server's client policy.

### Optional Local SPIRE Lab

The new [spire-setup.md](spire-setup.md) documents a separate local lab using official SPIRE `v1.15.5` images, trust domain `port.example.test`, a throwaway local root CA, short-lived join-token node attestation, and Docker-label workload registrations. It requires Docker socket and process/cgroup visibility for the agent. Treat its CA, token, selectors, and permissions as local-only. The main `docker/docker-compose.yml` must continue not to create SPIRE services.

Docker Desktop on macOS may not expose the Docker socket, process metadata, or Workload API Unix socket in the way the Linux container setup requires. Use a dedicated Linux Docker host/VM if the local lab cannot attest workloads. Do not work around this by disabling SPIFFE or mTLS.

## 7. Configuration

Validate startup configuration and never print credentials. Configuration includes:

- Server listen address: `PORT_SERVER_LISTEN` / `--listen`.
- Client target: `PORT_SERVER_URL` / `--server`.
- SPIFFE switch: `PORT_SPIFFE_ENABLED` / `--spiffe` (explicit flag wins).
- Workload API endpoint: `SPIFFE_ENDPOINT_SOCKET`.
- Expected server ID: `PORT_SERVER_SPIFFE_ID` / `--server-id`.
- Server policy: `PORT_AUTH_POLICY` / `--policy`.
- Collector interval, output directory, loopback listener, and `PORT_LOCAL_READER_SPIFFE_IDS`.

SPIFFE-off mode must reject non-loopback targets/listeners and use HTTP only for local development. The client must reject HTTP targets in SPIFFE mode.

## 8. Containers and Tooling

Use separate `docker/Dockerfile.server` and `docker/Dockerfile.client`, multi-stage static builds, non-root runtime UID/GID 65532, and the pinned Chainguard `static` image. Never place secrets/SVIDs in image layers. The images are independently buildable and the client can target a remote server.

Main app Compose configuration is in `docker/docker-compose.yml`; it requires an external SPIRE Agent socket and policy, does not deploy SPIRE, and should start the server independently from one-shot client requests. `make compose-up` starts the server; `make compose-client ARGS="query arrivals"` runs the client after the server is ready.

Use the existing Makefile targets and pinned tools. As of this revision: Go 1.27.3, go-spiffe v2.9.0, Vacuum v0.35.4, golangci-lint v2.15.0, govulncheck v1.1.4, and gosec v2.30.0. Maintain compatible patched indirect dependencies. No CI/CD in this iteration.

## 9. Tests and Replication

To reproduce code quality checks from the repository root:

```sh
go version
make install-tools
make check
+```
+
+To validate the Compose model without starting services, provide placeholder values:
+
+```sh
SPIFFE_SOCKET=/tmp/spire-agent.sock \
PORT_AUTH_POLICY_FILE=/absolute/path/to/authorization-policy.json \
PORT_SERVER_SPIFFE_ID=spiffe://port.example.test/ns/demo/sa/server \
docker compose -f docker/docker-compose.yml config --quiet
+```
+
+Build container images with `make docker-build-server docker-build-client` when Docker and Chainguard registry access are available. A local loopback smoke test may use `PORT_SPIFFE_ENABLED=false`, but must remain on loopback and is not evidence of SPIFFE integration.
+
+Unit/API and race tests must run without SPIRE. Actual Workload API socket access, node/workload attestation, TLS identity verification, real SVID/bundle rotation, and authorization against SPIRE require the local lab or an operator-provided SPIRE deployment. Report only tests executed; do not describe a successful compile as operational SPIRE validation.
+
+## 10. Current Verification State
+
+As of 2026-10-26, `make check` passed on Go 1.27.3, including format, vet, lint, unit/API tests, race tests, OpenAPI validation, `govulncheck`, `gosec`, and independent server/client builds. Vacuum reported zero OpenAPI errors and non-failing style/example warnings. A native loopback SPIFFE-off smoke test exercised the server, one-shot client, collector CSV writing, and latest-snapshot JSON GET. Compose configuration validation passed.
+
+The operational SPIRE path has not been integration-tested. Docker Engine responds, but image builds currently fail while pulling the pinned Chainguard base with a registry/blob-store `403 Forbidden`; do not claim image-build verification. Resolve registry access and rerun both Docker builds. Docker Desktop's SPIRE socket/attestation compatibility also remains unverified.
+
+## 11. Git and Owner Review
+
+The owner performs staging and commits. Do not commit, stage, rewrite history, push, or publish. Keep `.gitignore`, fictional-data disclaimers, safe policy examples, and secret-free exports. Review the current `git status` before proposing commit content so user changes remain included and no generated/local files are accidentally added.
+
+Suggested Conventional Commit boundaries for the work completed to date:
+
+- `feat(api): add fictional port information service`
+- `feat(client): add query collector and CSV snapshots`
+- `feat(security): add SPIFFE workload identity authorization`
+- `build(docker): add Chainguard server and client images`
+- `docs: document SPIRE lab setup and API examples`
+
+No commits are authorized by this prompt. At any future pre-commit review, report actual test results, Docker/Compose status, real SPIRE tests performed/not performed, limitations, secret review, and proposed commit boundaries.
