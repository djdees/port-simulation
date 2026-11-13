# ADR 005: Container Deployment and External SPIRE Dependency

**Status:** Accepted  
**Date:** 2026-10-26

## Context

The project requires independent container images for the server and client, with
Compose support. SPIRE infrastructure is managed externally.

## Decision

- **Separate images**: `docker/Dockerfile.server` and `docker/Dockerfile.client` are
  independent multi-stage builds. Each produces a minimal binary image.
- **Chainguard static runtime**: The runtime layer uses `cgr.dev/chainguard/static`,
  which provides a distroless base with no shell, package manager, or unnecessary
  libraries. This minimizes the attack surface.
- **Non-root user**: Both images run as UID/GID 65532 (Chainguard nonroot convention).
- **No secrets in layers**: SVIDs, private keys, and policy files are never baked into
  image layers. They are injected at runtime via bind mounts or environment variables.
- **External SPIRE**: The Compose file does not deploy SPIRE. It requires the caller to
  provide a Workload API socket path (`SPIFFE_SOCKET`) from an externally running SPIRE
  Agent. This keeps SPIRE lifecycle management separate from the application.
- **Compose profiles**: The client uses a `client` profile so it is not started by
  default with `compose-up`. It is run explicitly as a one-shot command.
- **Platform notes**: Docker Desktop on macOS may not expose the Docker socket or
  process metadata needed for SPIRE workload attestation. A dedicated Linux Docker host
  or VM is recommended for SPIRE integration testing.

## Consequences

- Chainguard registry access is required to pull the runtime base image. If the
  registry is unavailable (e.g., 403 response), image builds fail. This was an open
  issue as of 2026-10-26.
- The images are independently buildable and do not share a base layer with SPIRE
  components.
- Pinning the Chainguard digest prevents unintended runtime image updates.
