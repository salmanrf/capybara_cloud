---
paths:
  - "apps/**"
---

# Testing

## Apps: Back-end

### Structure

#### Integration Tests

Every business domain in `apps/backend` must have an integration test file at `apps/backend/tests/<domain>_integration_test.go`.

These tests cover the HTTP flow only:

- Permissions: 401 without the `sid` cookie, plus role and ownership checks.
- Routing, request validation (400/422, 413) and status codes.
- Interaction with the domain services: the right service method is called the expected number of times (stub call counters), and the handler maps service returns and errors to the right responses.

They must **not** test service logic itself. Services are stubbed here, and their logic is covered by [Domain Logic Tests](#domain-logic-tests).

The reference example is `apps/backend/tests/application_deployment_integration_test.go`:

- Mount the domain's router (`routes.Setup<Domain>Router`) on a `chi` mux over stub services from `tests/stubs.go`, and drive it with `ServeHTTP` and `httptest`.
- Write table-driven `t.Run` cases. Each case calls `stub.Clear()` in a `defer`.
- Compare values using `got_*`/`want_*` naming.
- No database.

#### Domain Logic Tests

These are unit tests for a domain's services. They live next to the service in `apps/backend/internal/<domain>/`, for example `internal/deployment/service_test.go`.

- Build the real service with its constructor (`NewService(...)`). Pass stubs for every dependency: repositories, other domains' services, Docker and similar. Stubs live in the package's `stubs.go` and `*_stubs.go` files, and other packages' stubs are reused, e.g. `application.StubApplicationService`.
- Assert on the service's return values and errors, and on its effects on the stubs (call counts and call args).
- When a service has many methods, a large one can get its own file, e.g. `service_deploy_build_test.go`, `service_deploy_push_test.go`.
- Set config through `config.GetConfig()` / `config.SetConfig(cfg)`, and use `/tmp/...` for filesystem paths.
- No HTTP and no database.
