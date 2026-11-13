# ADR 004: Dataset Storage and Embedding

**Status:** Accepted  
**Date:** 2026-10-26

## Context

The API must serve a deterministic, fictional dataset without requiring a database,
external file, or runtime data fetch. The dataset must be version-controlled and
validated.

## Decision

- **Embedded JSON**: The dataset is stored as `internal/server/data/port.json` and
  embedded into the server binary using Go's `//go:embed` directive. The server
  requires no separately mounted data file at runtime.
- **Hand-authored**: The dataset is hand-authored, deterministic, and checked into the
  repository. No random generation, seed, or external API is involved.
- **Validated at startup**: `data.Load()` parses the JSON and runs `domain.Validate()`
  at server startup. If the dataset is malformed or fails cross-reference integrity
  checks, the server exits with a clear error. Tests also run validation.
- **No database**: SQLite and other databases are out of scope for this iteration. The
  dataset is small enough to hold in memory and is read-only.
- **Dataset modification**: Edit `port.json`, then rebuild the server (`make
  build-server` or `make build`). The server validates on next startup. All IDs must
  remain stable for existing API consumers.

## Consequences

- The binary includes the full dataset — a minor size increase that is acceptable for
  a demonstration service.
- Changes to the dataset require a rebuild, which is appropriate for a version-controlled
  static fixture.
- Database introduction would be a future architectural decision requiring its own ADR,
  schema design, and migration strategy.
