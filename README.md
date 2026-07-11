# Capybara Cloud

Capybara Cloud is an open source, self-hosted, "Bring your own cloud", AI-assisted Platform as a Service (PaaS) engine.

# Documentations

- **[System Design Doc](https://link.excalidraw.com/l/35TAAPVqrQL/1c4VCvhYhX5)**
  The same document is available at `docs/system-design-docs.excalidraw`.

# Project Structure

The project started as a monolithic app, with some new modules gradually migrated into their own
microservices.

To keep everything in one place, the project is distributed in a monorepo.

The standard structure for each microservice is as follows:

- `api` -> This is where Web interfaces should be defined (API routes, controllers, middlewares).
- `docs` -> Centralized documentations, each module could have it's own root `docs` as well.
- `internal` -> Internal business logic modules, each domain is a separate go `package`.
- `sql` -> Contains migrations and queries.
- `pkg` -> Shared modules (utils, helpers, common dto).
- `tests` -> Integration & E2E tests.

# Setup

## Installation

```
go mod download
go mod tidy
```

## Migration

```
goose postgres <postgres_uri> up -dir sql/schema
```

## Environment Variables

```
POSTGRES_URI=
API_PORT=
AUTH_JWT_SECRET=
# Max formdata size for deployment in bytes
MAX_DEPLOY_FORM_SIZE=
# Max deployment bundle size in bytes
MAX_DEPLOY_BUNDLE_SIZE=
```

# Development Status

## Space Module — Complete

- Organization, project, and application CRUD
- Session-based authentication
- Full integration test coverage (TDD)

## Deployment Module — In Progress

- Single-host application deployment via multipart form upload.
- Custom multipart/form-data mocker for testing..

## Provisioning Module — Planned

- Event based host registration.
- TCP connection pinger.
