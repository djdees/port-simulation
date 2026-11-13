# Project: New Orleans Port Information Service

## 1. Role and objective

Act as a senior Go engineer, software architect, and security-conscious application developer.

Design and implement a Go server and a separate Go client for a fictionalized commercial port based on New Orleans, Louisiana. The application will demonstrate realistic API design, maintainable Go code, sound security practices, containerized deployment, and integration with SPIFFE identities issued by an operational SPIRE deployment.

This is an engineering demonstration, not a production port-management system. Nevertheless, code quality, security, documentation, and deployment practices must reflect established Go community standards and relevant industry guidance.

The server and client are separate applications. Each must have its own executable entry point, build target, binary, and Docker image. They must be independently usable and capable of communicating over a network.

The primary objective is to learn and demonstrate SPIFFE/SPIRE integration while building a useful, recognizable information service. Avoid unnecessary simulation complexity.

## 2. Domain and simulated data

Model a small, fictional port operating in the general context of the lower Mississippi River and Gulf Coast.

Include the following domain concepts:

- River traffic: barges, towboats, tows, and barge fleets.
- Commercial shipping: vessels carrying agricultural products, manufactured goods, raw materials, petroleum products, and liquefied natural gas (LNG).
- International traffic involving vessels associated with multiple countries.
- Offshore servicing vessels supporting offshore oil and gas operations.
- Port facilities, including fictional terminals, berths, anchorages, and service facilities.
- Traffic events, including arrivals, departures, berth assignments, loading, unloading, servicing, and delays.
- Cargo records with plausible commodities, origins, destinations, and handling requirements.

Use realistic terminology and internally consistent relationships. Keep the initial dataset small enough to understand and remember during an ad-hoc demonstration.

All vessel records, identifiers, manifests, schedules, berth assignments, and operational events must be fictional. Clearly label the dataset as simulated. Do not use live vessel-tracking feeds, actual operational records, real cargo manifests, or customer information. Do not imply affiliation with or endorsement by the Port of New Orleans or any actual port operator.

### Dataset implementation

Generate a hand-authored, deterministic dataset as part of the project.

Store the dataset as version-controlled JSON and embed it into the server binary using Go's `embed` package. The server must not require a separately mounted data file or an external database at runtime.

Requirements:

- Include enough linked records to demonstrate each supported resource type.
- Use stable, memorable identifiers.
- Keep vessel, barge, tow, cargo, terminal, berth, and event references internally consistent.
- Validate required fields, unique IDs, enum values where applicable, and relevant cross-references.
- Test dataset integrity.
- Avoid random generation, external APIs, and manual data refreshes.
- Document how to modify the dataset and rebuild the server.

Do not introduce SQLite or another database in the initial implementation. Reconsider persistence only if future requirements introduce meaningful mutable state, transactions, or concurrent updates.

## 3. API design and OpenAPI

Implement a versioned HTTP API returning JSON.

Potential endpoints include:

- `GET /api/v1/vessels`
- `GET /api/v1/vessels/{id}`
- `GET /api/v1/arrivals`
- `GET /api/v1/departures`
- `GET /api/v1/berths`
- `GET /api/v1/barges`
- `GET /api/v1/tows`
- `GET /api/v1/cargo`
- `GET /api/v1/terminals`
- `GET /api/v1/health/live`
- `GET /api/v1/health/ready`

These routes are illustrative. Refine the resource model where appropriate, document the decisions, and avoid implementing endpoints without a clear use case.

Requirements:

- Consistent resource identifiers and JSON field naming.
- Appropriate HTTP methods and status codes.
- A consistent error-response schema that does not leak internal implementation details.
- Input validation and sensible request-size limits.
- Bounded pagination and useful filtering where appropriate.
- Predictable handling of missing resources, malformed parameters, unsupported methods, and empty collections.
- Clear distinction between liveness and readiness.
- No unnecessary exposure of infrastructure details through health endpoints.

### Canonical OpenAPI specification

Create `openapi/openapi.yaml` as the single canonical API contract.

Use a stable OpenAPI version supported by the selected tooling. Document all implemented endpoints, parameters, request and response schemas, status codes, error formats, security requirements, and representative examples.

Document the relationship between HTTP-level security requirements and SPIFFE-based mutual TLS. OpenAPI declarations describe the API contract; they do not implement TLS validation or authorization.

Requirements:

- Keep the specification consistent with the server implementation.
- Validate the specification using an automated check.
- Include a `make openapi-check` target.
- Ensure Insomnia requests and examples align with the documented contract.
- Avoid maintaining a second competing API specification.
- Reference the canonical specification from the README and relevant documentation.

## 4. Go server

Implement the server using idiomatic Go and the standard library wherever practical. Introduce third-party dependencies only when they provide clear value.

Requirements:

- Separate transport, domain logic, data access, configuration, authentication, and authorization concerns.
- Validate configuration at startup.
- Use context-aware request handling.
- Implement graceful shutdown on operating-system termination signals.
- Configure bounded request processing and appropriate HTTP server timeouts.
- Use structured logging with useful request context.
- Avoid logging secrets, private keys, sensitive credential material, or unnecessary request data.
- Handle cancellation and client disconnects appropriately.
- Avoid unnecessary global mutable state and speculative abstraction.
- Return safe, actionable errors without exposing implementation details.
- Support multiple authorized clients.

The server must build independently of the client and must not need the client binary or source to run.

## 5. Go client

Implement a separate Go command-line client.

Requirements:

- Consume the documented API contract.
- Support representative operations, including listing arrivals, retrieving vessel details, and filtering cargo.
- Allow the server endpoint to be configured explicitly.
- Obtain workload identity through the SPIFFE Workload API.
- Use mutually authenticated TLS for protected API requests.
- Validate responses and handle HTTP errors, timeouts, cancellation, and malformed JSON.
- Print readable results suitable for interactive demonstrations.
- Separate API access from credential acquisition sufficiently to support focused testing.
- Avoid unnecessary dependencies and global mutable state.

The client must build and run independently of the server deployment. It must be able to connect to a server hosted elsewhere, provided that network access, trust configuration, and authorization are correctly configured.

Do not require the client to start a local server or assume that both binaries run on the same host.

## 6. SPIFFE/SPIRE integration

SPIFFE/SPIRE integration is a core architectural requirement, not an optional enhancement.

Use maintained, compatible Go libraries for the SPIFFE Workload API, X.509-SVID handling, and mutual TLS. Verify current APIs, supported versions, and credential-rotation behavior before implementation. Do not invent library interfaces or assume that an example from an outdated release remains valid.

Requirements:

- Obtain workload identities through the SPIFFE Workload API.
- Use X.509-SVIDs and the associated trust bundle for mutually authenticated TLS.
- Validate the server's SPIFFE identity and trust domain.
- Authenticate clients based on verified workload identities.
- Enforce explicit authorization policy based on verified SPIFFE IDs.
- Support credential rotation and trust-bundle updates through supported library mechanisms.
- Fail closed when required credentials, trust information, or authorization policy is unavailable.
- Never trust caller identity supplied in HTTP headers, query parameters, or request bodies.
- Never silently fall back to plaintext HTTP, insecure TLS verification, self-signed credentials, or unauthenticated access.
- Never log private keys or sensitive credential material.

Authentication and authorization are separate concerns. A valid workload identity must not automatically grant access to every resource or operation.

Document:

- The expected external SPIRE deployment.
- Trust domain and SPIFFE ID conventions.
- Required workload registrations and selectors.
- Workload API socket or endpoint configuration.
- Server and client identity expectations.
- Authorization policy and permitted operations.
- Required filesystem permissions and container access to the Workload API.
- Credential rotation and trust-bundle handling.
- Integration-test prerequisites.

Assume the operational SPIRE infrastructure is managed separately. Do not automatically deploy another SPIRE server or agent.

Any test-only identity mechanism must be isolated from operational mode, clearly documented, and incapable of being enabled accidentally in a normal deployment. Do not create a general-purpose insecure bypass to simplify local development.

## 7. Configuration and operational behavior

Use a small, explicit configuration surface, using environment variables or command-line flags where appropriate.

Document:

- Server listen address and port.
- Client target endpoint.
- Workload API endpoint or socket configuration.
- Expected server and client SPIFFE identities.
- Trust domain and authorization policy.
- Logging configuration.
- Health-check behavior.
- Any configuration required for Docker and Compose.

Validate configuration at startup and fail with actionable, non-sensitive errors.

Do not embed credentials, private keys, local machine-specific paths, or environment-specific assumptions in source code or images.

## 8. Observability

Provide useful operational visibility without introducing an unnecessarily large observability stack.

At minimum, include:

- Structured application logs.
- Useful request context and error reporting.
- Startup and shutdown messages.
- Liveness and readiness endpoints.
- Appropriate logging for authentication and authorization failures without exposing sensitive credential data.

Do not require an external logging or telemetry backend for the applications to run.

Create an ADR under `adr/` documenting the observability design, including logging choices, health checks, the initial scope, trade-offs, and possible future extensions.

## 9. Repository layout

Use conventional Go repository practices. Begin with the following structure, adapting package placement only where justified by Go conventions and package ownership.

```text
port-service/
├── cmd/
│   ├── server/
│   │   └── main.go
│   └── client/
│       └── main.go
├── internal/
│   ├── client/
│   ├── domain/
│   └── server/
│       └── data/
│           └── port.json
├── openapi/
│   └── openapi.yaml
├── adr/
├── docs/
├── insomnia/
├── docker/
│   ├── Dockerfile.server
│   ├── Dockerfile.client
│   └── docker-compose.yml
├── .gitignore
├── Makefile
├── README.md
├── go.mod
└── go.sum
```

Directory responsibilities:

- `cmd/`: executable entry points.
- `internal/`: private application implementation and domain logic.
- `openapi/`: canonical API contract.
- `adr/`: architectural decision records.
- `docs/`: supporting documentation other than the root README.
- `insomnia/`: importable API testing collections and safe environment templates.
- `docker/`: separate Dockerfiles and Compose configuration.
- Root files: Go module, Makefile, README, and `.gitignore`.

Avoid empty directories, unnecessary public packages, redundant abstractions, and generated files that do not serve a clear purpose.

## 10. Architectural decision records

Use a consistent ADR format containing:

- Title and status.
- Context and problem statement.
- Decision.
- Consequences and trade-offs.

Include ADRs for at least:

1. API and data-model architecture.
2. SPIFFE/SPIRE integration and the workload trust model.
3. Observability and operational diagnostics.
4. Dataset storage and embedding, if not adequately covered by another ADR.
5. Container deployment and the external SPIRE dependency, if the decision warrants a separate record.

Use sequential, meaningful filenames. Record decisions actually made by the project; do not invent retrospective justifications.

## 11. Docker and Compose

Provide two independent application images:

- Server image containing the server binary.
- Client image containing the client binary.

Each application must have its own Dockerfile under `docker/`, and the Compose file must be located at `docker/docker-compose.yml`.

Requirements:

- Use multi-stage builds.
- Use minimal runtime images where practical.
- Run as a non-root user.
- Minimize runtime packages and privileges.
- Do not place private credentials or SVIDs in image layers.
- Keep the server and client independently buildable.
- Support running the server without the client.
- Support running the client against a server running elsewhere.
- Support running both through Compose on a private application network.
- Make it straightforward to execute a one-shot client request.
- Document configuration, ports, networking, and startup requirements.

The operational SPIRE deployment is external to this project.

Document how the containers obtain access to the appropriate SPIRE Agent Workload API socket or another supported endpoint. Explain the required workload selectors and registration conditions. Do not assume that host socket mounting behaves identically on every operating system or container runtime.

The Compose file must not deploy an additional SPIRE server or agent unless explicitly approved as a future enhancement.

Container networking and SPIFFE identity are different concepts. A reachable IP address or Docker service name does not establish workload identity. TLS verification and authorization must remain enabled.

Do not introduce an insecure fallback merely to make the Compose example work. If a particular deployment environment cannot provide a supported Workload API connection, document the limitation and the appropriate integration setup.

## 12. Makefile and development tooling

Provide a standard, well-organized Go Makefile appropriate for this project.

Include conventional targets as appropriate:

- `help`
- `fmt`
- `vet`
- `lint`
- `test`
- `test-race`
- `build`
- `run-server`
- `run-client`
- `openapi-check`
- `security`
- `check`
- `clean`
- Docker build and Compose targets.

Build the server and client independently. The aggregate quality target should run the relevant formatting, analysis, test, OpenAPI, and build checks.

Targets must propagate failures, avoid destructive side effects, and clearly distinguish required tools from optional tools.

Document external development-tool prerequisites and installation methods. Avoid undocumented dependencies on local binaries, credentials, or environment variables.

Use appropriate maintained tooling, including `govulncheck` and relevant static analysis/security tools where practical. Clearly manage tool versions.

Do not add CI/CD configuration in this iteration. The Makefile is the initial quality-gate interface. If CI/CD is introduced later, its configuration and operational documentation will be handled as a separate enhancement.

## 13. Testing strategy

Include meaningful unit, API, and integration tests.

Cover:

- Domain validation and dataset integrity.
- API success responses and error cases.
- Filtering, pagination, and input validation where implemented.
- Client response decoding and failure handling.
- Authentication and authorization.
- Invalid workload identities and TLS trust failures.
- Relevant configuration, startup, and shutdown behavior.
- Race detection.
- Docker build validation where the required tooling is available.

Separate tests that run without SPIRE from integration tests requiring a real SPIRE deployment.

Provide documented instructions for validating actual Workload API use, SVID validation, mutual TLS, rotation behavior, and authorization against SPIRE.

Do not claim integration has been verified against an operational SPIRE deployment unless those tests have actually been performed.

## 14. Insomnia artifacts

Provide an importable Insomnia collection under `insomnia/`.

Include representative requests for:

- Listing vessels, arrivals, and departures.
- Retrieving a vessel or barge by ID.
- Filtering cargo.
- Listing terminals and berths.
- Missing-resource and invalid-parameter cases.
- Authentication-related cases where supported by the test setup.

Keep the collection aligned with `openapi/openapi.yaml`.

Document how to configure Insomnia for ordinary API exploration and the limitations of using it against SPIFFE-protected endpoints. Do not assume Insomnia automatically obtains workload identities through the SPIFFE Workload API.

Do not include private keys, actual credentials, or machine-specific secrets in exported environments.

## 15. README and supporting documentation

The root `README.md` must provide a practical introduction covering:

- Purpose, scope, and simulated-data disclaimer.
- Architecture overview and a simple diagram where useful.
- Repository layout.
- Prerequisites and supported Go version.
- Native build and run instructions.
- Independent server and client usage.
- Makefile targets.
- Configuration reference.
- Docker images and Compose usage.
- Example API requests and responses.
- Link to the canonical OpenAPI specification.
- Insomnia setup.
- SPIFFE/SPIRE prerequisites and architecture.
- Workload registration and authorization setup.
- Testing and security checks.
- Troubleshooting and known limitations.

Use `docs/` for detailed setup procedures, architectural explanations, and operational material beyond the README.

Ensure that documentation matches the implementation. Do not document nonexistent endpoints, commands, configuration options, or capabilities.

## 16. Git and repository hygiene

The repository owner performs all Git commits.

Generate a suitable `.gitignore` that excludes, as appropriate:

- Compiled binaries and build artifacts.
- Test outputs and coverage files.
- Temporary files.
- Local environment and configuration files.
- Secrets and private credentials.
- IDE- and editor-specific artifacts.
- Other generated files that should not be version controlled.

Do not ignore source files, the canonical OpenAPI specification, the simulated dataset, Insomnia collection exports, ADRs, or required Docker artifacts.

All commits made by the repository owner will follow Conventional Commits. At the end of the work, propose meaningful commit boundaries and conventional commit messages.

Do not create commits, amend or rewrite history, push changes, or publish artifacts. Do not stage files unless explicitly requested.

## 17. Required workflow and approval gates

### Gate A: Standards review

Before implementation, review the specification and identify missing engineering, security, repository, documentation, and workflow standards.

Present proposed additions or unresolved decisions. Wait for approval before treating new preferences as permanent requirements.

### Gate B: Architecture review

Propose and review:

- Repository tree.
- Domain model and sample dataset.
- API resource model and OpenAPI structure.
- Dependency and tool choices.
- SPIFFE trust and authorization design.
- Credential lifecycle and rotation behavior.
- Container architecture and Workload API access.
- Testing strategy.
- Initial ADRs.

Explain significant trade-offs and surface assumptions. Wait for approval before making major architectural commitments.

### Gate C: Implementation and verification

Implement the approved design.

Run the applicable formatting, test, race, vet, lint, build, OpenAPI, and security checks. Report actual results, skipped checks, and environmental limitations.

Do not claim a check passed if it was not executed. Do not claim operational SPIRE integration has been verified without a real integration test.

### Gate D: Pre-commit review

Do not create Git commits.

Present:

- Final repository tree and changed-file summary.
- Build and test results for both applications.
- Lint, OpenAPI, and security-check results.
- SPIRE integration tests performed and not performed.
- Docker build results and Compose limitations.
- Known issues and unresolved decisions.
- Confirmation that `.gitignore` exists.
- Confirmation that no secrets, credentials, or unrelated files are included.
- Proposed Conventional Commit messages and logical commit boundaries.

Leave the repository ready for review and let the repository owner decide when and how to commit.

## 18. Definition of done

The project is ready for owner review when:

1. The server and client build into separate Go binaries.
2. Each application has an independent Docker image.
3. Both applications can be built and operated independently.
4. The client can connect to a remote server without requiring a local server container.
5. Compose supports convenient combined use without deploying duplicate SPIRE infrastructure.
6. The server embeds a deterministic, validated, fictional dataset.
7. The API is documented in `openapi/openapi.yaml`.
8. SPIFFE-based mutual TLS and explicit identity-based authorization are implemented.
9. Credential rotation and trust-bundle updates are handled using supported mechanisms.
10. Meaningful unit, API, and integration tests exist.
11. The Makefile provides repeatable development and quality checks.
12. README, OpenAPI, ADRs, Insomnia artifacts, Docker configuration, and supporting documentation agree with the implementation.
13. `.gitignore` is present and the repository contains no committed credentials.
14. Actual verification results and environmental limitations are documented.
15. The final review includes proposed Conventional Commit messages.
16. No Git commits have been created.

Prioritize correctness, security, maintainability, and a credible SPIRE integration over feature count. When requirements are ambiguous, surface the ambiguity instead of silently inventing a complex solution.
