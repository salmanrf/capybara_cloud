# Capybara Cloud

Capybara Cloud is an open source, self-hosted, "Bring your own cloud", AI-assisted Platform as a Service (PaaS) engine.

# Documentations

- **[System Design Doc](https://link.excalidraw.com/l/35TAAPVqrQL/1c4VCvhYhX5)**
  The same document is available at `docs/system-design-docs.excalidraw`.
- **Bruno collection** for the REST API lives in `apps/backend/bruno`.

# Project Structure

The project started as a monorepo app, with some new modules gradually migrated into their own
microservices.

To keep everything in one place, the project is distributed in a monorepo, using basic nx-monorepo
(`./nx` wrapper script, requires Node.js).

Monorepo Structure:

- `apps/` -> Deployable applications (frontends, backends, CLIs)
  - `apps/backend` -> Go API server (chi + pgx)
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

## Installation

```
cd apps/backend
go mod download
go mod tidy
```

## Database & Migration

Local database runs on Supabase CLI. Migrations live in `apps/backend/supabase/migrations`.

```
cd apps/backend
supabase start
supabase db reset   # applies all migrations
```

## Code Generation

Queries are compiled with sqlc into `packages/shared-go/database` (gitignored, must be generated).

```
cd apps/backend
sqlc generate
```

## Build & Run

```
cd apps/backend
./run.sh dev    # or: ./run.sh prod
```

`run.sh` builds `bin/backend` via `make dev|prod` and runs it with `ENV` set. Logs are written to
`apps/backend/apps.backend.logs`.

## Environment Variables

Loaded from `apps/backend/.env`.

```
# Required
POSTGRES_URI=
API_PORT=
AUTH_JWT_SECRET=
AUTH_JWT_ISSUER=
AUTH_JWT_AUDIENCE=
DOCKER_REGISTRY=
DOCKER_NAMESPACE=
# Max formdata size for deployment in bytes
MAX_DEPLOY_FORM_SIZE=
# Max deployment bundle size in bytes
MAX_DEPLOY_BUNDLE_SIZE=

# Optional
DOCKER_USER=
DOCKER_ACCESS_TOKEN=
BASE_ARTIFACT_PATH=
BASE_BUILD_PATH=
# Defaults to internal/deployment/templates
DOCKER_TEMPLATES_DIR=
```

# Development Status

**Conventions**

- **Module**: modules in the context of this document refers not to specifically go modules but a general encapsulation of related pieces of code.

## Space Module

- User, organization, project, and application CRUD
- JWT-based authentication
- Full integration test coverage (TDD)

## Deployment Module

Multi-step deployment module: extract bundle -> build Docker image -> push to registry -> start.
Deployment instances track each step; a listener service handles step transitions and errors.

## Masbro Worker Module

This is the core of Capybara Cloud, Masbro Workers are the services that actually runs the deployed applications and providing interfaces to them.

Current state: `internal/masbro-worker` runs a deployment instance as a Docker container from the
pushed image.

## Masbro Manager Module

Not started.

## Observability

Global structured logger (`slog`) in `packages/shared-go/logger`, request logging middleware.
