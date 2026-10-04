# Capybara Cloud

Capybara Cloud is an open source, self-hosted, "Bring your own cloud", AI-assisted Platform as a Service (PaaS) engine.

# Documentations

- **[System Design Doc](https://link.excalidraw.com/l/35TAAPVqrQL/1c4VCvhYhX5)**
  The same document is available at `docs/system-design-docs.excalidraw`.
- **Bruno collection** for the REST API lives in `apps/backend/bruno`.

# Project Structure

The project started as a monorepo app, with some new modules gradually migrated into their own
microservices.

To keep everything in one place, the project is distributed in an Nx monorepo. Nx is pinned as a
dev dependency in the root `package.json` and run through a globally installed `nx` (see Setup).

Monorepo Structure:

- `apps/` -> Deployable applications (frontends, backends, CLIs)
  - `apps/backend` -> Go API server (chi + pgx)
  - `apps/frontend` -> Web dashboard (React + TypeScript, Vite, TanStack Router + Query)
- `packages/` -> Shared libraries consumed by apps or other packages
  - `packages/shared-go` -> Shared Go module: `database` (sqlc generated, gitignored), `deployment`,
    `docker`, `errors`, `logger`, `utils`, `tests`

The standard structure for each microservice is as follows:

- `api` -> This is where Web interfaces should be defined (API routes, handlers, middlewares).
- `docs` -> Centralized documentations, each module could have it's own root `docs` as well.
- `internal` -> Internal business logic modules, each domain is a separate go `package`.
- `sql` -> sqlc queries (`sql/queries`).
- `supabase` -> Local Supabase setup and database migrations (`supabase/migrations`).
- `pkg` -> Shared modules (utils, helpers, common dto).
- `tests` -> Integration & E2E tests.
- `bin` -> Build output (gitignored).

# Setup

Every project is registered with Nx (`apps/backend`, `apps/frontend`, `packages/shared-go`), so all
commands below run from the repository root with `nx`. List the projects with `nx show projects`,
and a project's targets with `nx show project <name>`.

## Installation

Requires Node.js, pnpm, Go, the Supabase CLI and sqlc.

```
npm install -g nx            # once per machine: puts `nx` on your PATH
pnpm install                 # from the repo root: installs Nx, the frontend and the Go modules
```

The global `nx` delegates to the version pinned in the root `package.json`, so everyone runs the
same Nx regardless of which global version they installed. If `nx` reports "Could not find Nx
modules", run `pnpm install` in the repository root.

The root `pnpm-workspace.yaml` makes `apps/frontend` a workspace package with a single
`pnpm-lock.yaml` at the root, so always install from the root.

## Code Generation

Queries are compiled with sqlc into `packages/shared-go/database` (gitignored, must be generated).

```
nx sqlc backend            # sqlc generate
```

`build`, `dev`, `serve` and `test` on the backend (and `test` on shared-go) depend on `sqlc`, so
Nx runs it first. Its result is cached until `sqlc.yaml`, `sql/` or the migrations change.

## Build & Run

```
nx run-many -t dev -p backend frontend   # run backend and frontend together

nx dev backend             # ./run.sh dev  (debug build, ENV=development)
nx serve backend           # ./run.sh prod (production build, ENV=production)
nx build backend           # make dev -> apps/backend/bin/backend
nx build backend -c production           # make prod
```

`run.sh` builds `bin/backend` via `make dev|prod` and runs it with `ENV` set. Logs are written to
`apps/backend/apps.backend.logs`.

## Tests

```
nx test backend            # go test ./... in apps/backend
nx test shared-go          # go test ./... in packages/shared-go (docker e2e needs a Docker daemon)
nx test frontend           # vitest
nx run-many -t test        # everything
```

To pass flags through, append them: `nx test backend -- -run TestAuthSignupIntegration`.

## Backend

Located in `apps/backend`. Database, migrations and environment variables are documented in
[`apps/backend/README.md`](apps/backend/README.md).

```
pnpm install                 # Nx and the Go modules
nx supabase-reset backend    # apply migrations to the local database
nx dev backend               # starts Supabase, runs sqlc, then the API on API_PORT
```

## Frontend

```
nx dev frontend            # http://localhost:5173, proxies /api to the backend
nx test frontend           # api module tests
nx typecheck frontend
nx build frontend          # static output in apps/frontend/dist/
```

The backend authenticates with a `SameSite=Strict`, `HttpOnly` `sid` cookie, so the frontend must
reach the API on its own origin. In dev, Vite proxies `/api` to `http://localhost:8888` (override
with `API_URL`); in production, serve `dist/` and `/api` behind the same host.

All backend calls go through `src/api` (`createApi()`), which owns URLs, cookies and the response
envelope, and throws `ApiError` on failure. Routes under `src/routes/_authed/` require a session.
