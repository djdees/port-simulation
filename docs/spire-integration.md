# SPIRE Integration Reference

This document describes the external SPIRE deployment requirements for the New Orleans
Port Information Service. The application does not deploy SPIRE — it consumes an
existing SPIRE Agent's Workload API.

## Trust Domain

```
port.example.test
```

All SPIFFE IDs follow the form `spiffe://port.example.test/<path>`.

## Expected SPIFFE IDs

| Component | SPIFFE ID |
|---|---|
| Server | `spiffe://port.example.test/ns/demo/sa/server` |
| Client | `spiffe://port.example.test/ns/demo/sa/client` |

These are conventional. Adjust the path components and authorization policy to match
your actual SPIRE workload registrations.

## Required Workload Registrations

Register at minimum two workloads with your SPIRE Server:

```sh
# Server registration
spire-server entry create \
  -spiffeID spiffe://port.example.test/ns/demo/sa/server \
  -parentID  spiffe://port.example.test/ns/demo/node1 \
  -selector  docker:label:com.port-simulation.role:server

# Client registration
spire-server entry create \
  -spiffeID spiffe://port.example.test/ns/demo/sa/client \
  -parentID  spiffe://port.example.test/ns/demo/node1 \
  -selector  docker:label:com.port-simulation.role:client
```

Selector types depend on your attestation method (Docker labels, Kubernetes SATs,
Unix process UIDs, etc.). Adjust for your environment.

## Workload API Socket

The SPIRE Agent exposes the Workload API on a Unix domain socket. Configure the
socket path in both the server and client:

```
SPIFFE_ENDPOINT_SOCKET=unix:///run/spire/sockets/agent.sock
```

The standard socket path in Docker is `/run/spire/sockets/agent.sock`, bind-mounted
into the container from the host. See `docker/docker-compose.yml`.

## Authorization Policy File

The server reads a JSON authorization policy at startup. Example:

```json
{
  "clients": [
    {
      "spiffe_id": "spiffe://port.example.test/ns/demo/sa/client",
      "permissions": ["port:read"]
    }
  ]
}
```

Set `PORT_AUTH_POLICY` or `--policy` to the path of this file.
**Do not commit a policy file with real SPIFFE IDs to the repository.**

## Credential Rotation

go-spiffe keeps the `X509Source` alive for the server process lifetime. When SPIRE
rotates the SVID or trust bundle, go-spiffe fetches the new material automatically and
updates the TLS configuration without a server restart.

## Integration Test Prerequisites

To run integration tests against a real SPIRE deployment:

1. A SPIRE Server and Agent must be running and healthy.
2. Both server and client workloads must be registered with appropriate selectors.
3. The Workload API socket must be accessible from the test environment.
4. The agent must successfully attest the workloads before the test.

Standard `make test` runs unit and API tests that do not require SPIRE.
SPIRE integration tests are opt-in and must be executed separately with a live
environment.

## Platform Notes

- **Linux**: SPIRE Agent natively supports Docker socket and cgroup/process selectors.
- **macOS (Docker Desktop)**: The Docker socket and process metadata may not be exposed
  to the Linux container environment in the way SPIRE expects. Use a dedicated Linux
  VM or Docker host for SPIRE integration testing on macOS.
- **Kubernetes**: Use SPIRE Kubernetes node attestation and pod identity selectors
  (see official SPIRE Kubernetes quickstart).
