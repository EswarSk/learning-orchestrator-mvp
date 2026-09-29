# Go services

This is **one Go module containing five independently started programs**, not one backend server. `compose.yaml` builds the same Dockerfile five times with a different `SERVICE` argument and runs each resulting binary in its own container.

| Entrypoint | Implementation | Job |
|---|---|---|
| `cmd/experience` | `internal/experience` | Public mobile API, identity, consent, devices |
| `cmd/learning` | `internal/learning` | Tracks, courses, practice, assessment, evidence |
| `cmd/context` | `internal/context` | Location/Calendar signals and Temporal dispatch |
| `cmd/opportunity` | `internal/opportunity` | Opportunity inbox and actions |
| `cmd/opportunity-worker` | `internal/opportunity` | Temporal decisions and notification outbox |

`internal/platform` contains shared HTTP, PostgreSQL, and Temporal setup; `internal/contracts` contains the worker's internal trigger type. `migrations/` changes the active database. The service code is grouped in one module for atomic local changes and one dependency graph; deployment boundaries are the five binaries, not the Go module boundary.

The services currently share one PostgreSQL database with domain schemas and some cross-schema queries. They are separately deployed processes, but **not yet independently owned databases**.
