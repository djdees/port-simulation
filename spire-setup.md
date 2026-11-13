# Optional Local SPIRE Lab Guide

> **WARNING — LOCAL DEV ONLY.**  
> This guide creates a throwaway SPIRE deployment for local testing. The CA, tokens,
> selectors, and permissions produced here are local-only and must never be used in
> staging or production. This is not production SPIRE guidance.

This optional guide sets up a disposable SPIRE Server and Agent on a Linux Docker host
to exercise the SPIFFE identity path without an operational SPIRE deployment.

## Prerequisites

- Docker Engine on Linux (not Docker Desktop on macOS — see note below)
- Docker socket and cgroup/process visibility for the SPIRE Agent
- SPIRE v1.15.5 images

> **macOS note:** Docker Desktop on macOS does not expose the Docker socket or process
> metadata in a way that supports SPIRE's Linux-native workload attestation. Use a
> dedicated Linux VM or bare-metal Linux host for this lab.

## 1. Create a Docker network

```sh
docker network create spire-lab
```

## 2. Start the SPIRE Server

```sh
docker run -d --name spire-server \
  --network spire-lab \
  -p 8081:8081 \
  ghcr.io/spiffe/spire-server:1.15.5 \
  -config /opt/spire/conf/server/server.conf
```

The default server config uses an in-memory SQLite database and an in-process CA
suitable for local testing only.

## 3. Bootstrap a join token

```sh
# Generate a one-time join token
TOKEN=$(docker exec spire-server \
  /opt/spire/bin/spire-server token generate \
  -spiffeID spiffe://port.example.test/ns/demo/node1 \
  | awk '/Token:/ {print $2}')
echo "Join token: $TOKEN"
```

## 4. Start the SPIRE Agent

```sh
docker run -d --name spire-agent \
  --network spire-lab \
  --pid host \
  -v /var/run/docker.sock:/var/run/docker.sock \
  -v /run/spire/sockets:/run/spire/sockets \
  ghcr.io/spiffe/spire-agent:1.15.5 \
  -config /opt/spire/conf/agent/agent.conf \
  -joinToken "$TOKEN"
```

The agent requires `--pid host` and Docker socket access for Docker-label workload
attestation.

## 5. Register workloads

```sh
# Server workload
docker exec spire-server \
  /opt/spire/bin/spire-server entry create \
  -spiffeID spiffe://port.example.test/ns/demo/sa/server \
  -parentID  spiffe://port.example.test/ns/demo/node1 \
  -selector  docker:label:com.port-simulation.role:server \
  -ttl 3600

# Client workload
docker exec spire-server \
  /opt/spire/bin/spire-server entry create \
  -spiffeID spiffe://port.example.test/ns/demo/sa/client \
  -parentID  spiffe://port.example.test/ns/demo/node1 \
  -selector  docker:label:com.port-simulation.role:client \
  -ttl 3600
```

## 6. Label application containers

When running the server and client containers, add the matching Docker labels:

```sh
docker run --label com.port-simulation.role=server ... port-server:latest
docker run --label com.port-simulation.role=client ... port-client:latest
```

## 7. Mount the Workload API socket

Bind-mount the socket into the application containers:

```
-v /run/spire/sockets/agent.sock:/run/spire/sockets/agent.sock
```

Then set:

```
SPIFFE_ENDPOINT_SOCKET=unix:///run/spire/sockets/agent.sock
```

## 8. Create an authorization policy

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

Save this as `local-policy.json` (do not commit it).

## 9. Teardown

```sh
docker rm -f spire-server spire-agent
docker network rm spire-lab
rm -rf /run/spire
```

## Verification Status

This lab guide has not been end-to-end verified in this workspace. Report actual test
results when performing SPIRE integration testing.
