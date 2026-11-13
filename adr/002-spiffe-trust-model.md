# ADR 002: SPIFFE/SPIRE Integration and Workload Trust Model

**Status:** Accepted  
**Date:** 2026-10-26

## Context

The project must demonstrate SPIFFE/SPIRE workload identity, mTLS, and explicit
authorization — not just TLS with a self-signed cert. The design must fail closed,
support credential rotation, and not introduce insecure fallbacks.

## Decision

- **go-spiffe v2**: Use `github.com/spiffe/go-spiffe/v2` for Workload API access,
  X.509-SVID/trust-bundle handling, mTLS helpers, and rotation support.
- **X509Source lifecycle**: The source is kept open (not fetched once and closed) so
  that go-spiffe can push updated SVIDs and trust bundles automatically as they rotate.
- **Server mTLS**: The server presents its X.509-SVID and requires clients to present
  a valid SVID from the same trust domain. The TLS layer accepts only client IDs in the
  authorization policy.
- **Authorization policy**: A JSON file maps exact client SPIFFE IDs to allowed
  permissions (`port:read`). Authentication (valid X.509-SVID) and authorization (policy
  lookup) are enforced separately. A valid identity alone does not grant access.
- **Never trust headers**: SPIFFE identity is never accepted from HTTP headers, query
  parameters, or request bodies.
- **SPIFFE-off mode**: An explicit `--spiffe=false` / `PORT_SPIFFE_ENABLED=false` flag
  enables plain-HTTP mode for local development. It rejects non-loopback addresses and
  logs a prominent warning. It is incapable of becoming a default or remote deployment
  mode.
- **No insecure fallback**: The server never falls back to plain HTTP if SPIFFE is
  enabled. Missing credentials cause startup failure, not silent degradation.

## Consequences

- SPIRE infrastructure must be external and pre-configured. The project does not deploy
  SPIRE in the main Compose file.
- An optional local SPIRE lab guide (spire-setup.md) documents a throwaway setup for
  development. It is explicitly labeled non-production.
- Integration tests for the SPIFFE path require a real SPIRE environment and are not
  part of the standard `make test` run.
- Credential rotation is handled transparently by go-spiffe without server restart.
