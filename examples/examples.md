# API Examples

All examples use the fictional New Orleans Port Information Service dataset.
In SPIFFE-off development mode, replace `https://localhost:8443` with
`http://localhost:8080`.

## Health

```sh
# Liveness
curl http://localhost:8080/api/v1/health/live

# Readiness
curl http://localhost:8080/api/v1/health/ready
```

## Vessels

```sh
# List all vessels
curl http://localhost:8080/api/v1/vessels

# Filter by kind
curl "http://localhost:8080/api/v1/vessels?kind=tanker"

# Filter by flag country
curl "http://localhost:8080/api/v1/vessels?flag_country=US"

# Pagination
curl "http://localhost:8080/api/v1/vessels?limit=2&offset=0"

# Get vessel by ID
curl http://localhost:8080/api/v1/vessels/VSL-001
```

## Arrivals and Departures

```sh
# List arrivals
curl http://localhost:8080/api/v1/arrivals

# Arrivals at a specific facility
curl "http://localhost:8080/api/v1/arrivals?facility_id=FAC-001"

# Arrivals in a time range
curl "http://localhost:8080/api/v1/arrivals?since=2026-10-20T00:00:00Z&until=2026-10-21T23:59:59Z"

# List departures
curl http://localhost:8080/api/v1/departures
```

## Barges

```sh
# List all barges
curl http://localhost:8080/api/v1/barges

# Filter by capability
curl "http://localhost:8080/api/v1/barges?capability=dry_bulk"

# Barges in a tow
curl "http://localhost:8080/api/v1/barges?tow_id=TOW-001"

# Get barge by ID
curl http://localhost:8080/api/v1/barges/BGE-001
```

## Tows

```sh
# List tows
curl http://localhost:8080/api/v1/tows

# Filter by towboat
curl "http://localhost:8080/api/v1/tows?towboat_id=VSL-005"

# Get tow by ID
curl http://localhost:8080/api/v1/tows/TOW-001
```

## Cargo

```sh
# List all cargo
curl http://localhost:8080/api/v1/cargo

# Filter by commodity
curl "http://localhost:8080/api/v1/cargo?commodity=petroleum"

# Filter by vessel
curl "http://localhost:8080/api/v1/cargo?vessel_id=VSL-001"

# Filter by barge
curl "http://localhost:8080/api/v1/cargo?barge_id=BGE-001"

# Get cargo by ID
curl http://localhost:8080/api/v1/cargo/CGO-001
```

## Facilities, Terminals, Berths

```sh
# List facilities
curl http://localhost:8080/api/v1/facilities

# List commercial facilities
curl "http://localhost:8080/api/v1/facilities?kind=commercial"

# Get facility by ID
curl http://localhost:8080/api/v1/facilities/FAC-001

# Terminals in a facility
curl "http://localhost:8080/api/v1/terminals?facility_id=FAC-001"

# Get terminal by ID
curl http://localhost:8080/api/v1/terminals/TRM-001

# Berths in a terminal
curl "http://localhost:8080/api/v1/berths?terminal_id=TRM-001"

# Get berth by ID
curl http://localhost:8080/api/v1/berths/BTH-001
```

## Client CLI

```sh
# SPIFFE-off mode (development only)
export PORT_SPIFFE_ENABLED=false
export PORT_SERVER_URL=http://localhost:8080

# One-shot queries
./bin/client query arrivals
./bin/client query departures
./bin/client query vessels
./bin/client get-vessel VSL-001
./bin/client filter-cargo commodity=petroleum

# Start the periodic collector
./bin/client collect

# Check a snapshot (while collector is running)
curl http://127.0.0.1:8081/api/v1/latest/arrivals
```

## Error Cases

```sh
# Not found
curl http://localhost:8080/api/v1/vessels/VSL-999
# → 404 application/problem+json

# Invalid filter value
curl "http://localhost:8080/api/v1/vessels?kind=zeppelin"
# → 400 application/problem+json

# Invalid pagination
curl "http://localhost:8080/api/v1/vessels?limit=-1"
# → 400 application/problem+json
```
