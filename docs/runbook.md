# Verification and release runbook

The active backend is five separate Go processes with PostgreSQL and Temporal. The former TypeScript backend and its pg-boss/RLS instructions are preserved under `legacy/typescript-backend/`, not the deployment path.

## Local checks

```sh
pnpm install --frozen-lockfile
docker compose up -d postgres
# Once per local machine: use a separate, disposable database for integration tests.
docker compose exec -T postgres createdb -U orchestrator orchestrator_test
cd backend
DATABASE_URL=postgres://orchestrator:orchestrator@localhost:54322/orchestrator_test go run ./cmd/migrate
TEST_DATABASE_URL=postgres://orchestrator:orchestrator@localhost:54322/orchestrator_test go test ./...
go vet ./...
cd ..
pnpm content:check
pnpm typecheck
pnpm --filter @orchestrator/mobile build
docker compose up -d --build
curl http://localhost:3000/health/ready
```

`TEST_DATABASE_URL` must name a database ending in `_test`; the integration test rejects any other target. The test uses unique users and removes its records. CI creates and migrates its own `orchestrator_test` database before running the same checks.

## Release gate

Do not give this build to live users yet. The following require implementation and evidence, not just a green CI run:

- AI-generated courses are permitted before educator review when clearly labeled. Live-provider, safety, and multi-subject quality evaluation are still required.
- Live OpenAI assessment quality and safety evaluation across multiple subjects; local tests already gate progress on assessed understanding and schedule review.
- Real OIDC sign-in, refresh, revocation, and account switching against the chosen provider.
- Calendar connector consent/sync/revocation, plus physical-iPhone geofence → offer → notification/voice checks and a real Expo/APNs receipt. Push tickets and receipts are locally mock-tested, not live-delivered.
- Account deletion, physical-device and large-account export checks, production role isolation, monitoring, backup restore, and rollback practice. Local account export and simulator sharing are verified. Paid subscriptions and background email access are deferred from the first launch.

Production deploy and rollback commands cannot be finalized until the hosting, identity, AI, connector, and billing providers are selected and tested. The same OCI services can run on AWS, Azure, or GCP; production needs TLS ingress, private service networking, managed PostgreSQL, Temporal, distinct credentials, secrets management, and observability. Never use the Docker Compose credentials or local auth token outside local development.
