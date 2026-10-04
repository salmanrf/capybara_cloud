# Project repository

Status: ready-for-agent

## Problem Statement

The project domain is the odd one out in the backend. The application and deployment domains reach the database through a `Repository` interface over sqlc `Queries`, so their services can be built in tests with a hand-written stub repository. The project service instead holds the Postgres connection pool and the sqlc `Queries` directly, and opens transactions itself.

For a developer this means:

- The project service has no domain logic tests at all, because it cannot be built without a real database.
- Its transaction handling is fragile and already buggy. Deleting a project defers the rollback before checking whether the transaction could be started, which panics on a nil transaction if the pool is unavailable. Creating a project ignores the commit error, so a failed commit is reported as a successful create.
- Upcoming project work (per-org listing, project members, deletion rules from the management-API discussion) would have to be written against the pool and tested only through the HTTP layer with the whole service stubbed out.

## Solution

Add a project repository, shaped like the application and deployment repositories, and make the project service depend on it instead of on the pool and `Queries`.

Statements that must succeed or fail together are exposed by the repository as single atomic methods, which own their transaction internally: creating a Project together with its owner membership, and deleting a Project together with its memberships. The service keeps the domain rules (the user must exist, "not found" is reported as no result, the creator becomes owner) and no longer imports any database driver package.

Behaviour seen through the API does not change. This is a refactor plus the two transaction bug fixes.

## User Stories

1. As a backend developer, I want the project service to depend on a repository interface, so that I can build the real service in a unit test without a database.
2. As a backend developer, I want the project repository to follow the same shape as the application repository (constructor taking the context and its dependencies, one method per database operation, params as sqlc param structs, rows returned as pointers), so that every domain reads the same way.
3. As a backend developer, I want creating a Project and its owner membership to be one repository method, so that the service cannot forget to commit or roll back.
4. As a backend developer, I want deleting a Project and its memberships to be one repository method, so that a Project is never left without its memberships or vice versa.
5. As a backend developer, I want the project service to stop importing the pgx and pgxpool packages, so that database concerns stay behind the repository.
6. As a backend developer, I want a stub project repository with configurable return values and errors, call counters and recorded call args, so that I can assert how the service uses the repository.
7. As a backend developer, I want a stub user service in the project package, so that I can control whether the creating user exists.
8. As a user creating a Project, I want the Project and my owner membership to be created together or not at all, so that I never end up with a Project I can't access.
9. As a user creating a Project, I want an error if the database fails to commit, so that I'm not told the Project was created when it wasn't.
10. As a user deleting a Project, I want a clean error rather than a server crash when the database is unreachable, so that the API stays up.
11. As a user deleting a Project, I want the Project and its memberships removed together or not at all, so that a failed delete leaves everything as it was.
12. As a user, I want project create, read, list, update and delete to behave exactly as before, so that the refactor is invisible to me.
13. As a user creating a Project with a name that already exists, I want the same "already exists" response as before, so that the duplicate-name handling is unchanged.
14. As a user fetching a Project I'm not a member of, or that doesn't exist, I want the same not-found response as before, so that access rules are unchanged.
15. As a user listing my Projects with none to show, I want an empty list as before, not an error.
16. As a backend developer, I want the application domain's existing project service stub to keep working unchanged, so that this refactor doesn't ripple into other packages.
17. As a backend developer, I want the HTTP integration tests to keep passing unchanged, so that I know the API contract is intact.

## Implementation Decisions

**New module: project repository**

- A `ProjectRepository` interface plus an unexported implementation in the project package, and a constructor `NewRepository`, mirroring the application repository.
- The constructor takes the context, the connection pool and the sqlc `Queries`. The pool is needed only by the atomic methods, to open their transactions.
- Methods take sqlc param structs (or a plain ID where the query takes a single UUID) and return pointers to sqlc rows plus an error, passing sqlc and pgx errors through unchanged. Translating errors, such as "no rows" into a nil result, stays in the service, as it does in the application domain.
- Interface:
  - `CreateProjectWithMember(project params, member role)`: in one transaction, inserts the Project, then inserts a project member row for the given user and role on the new Project. Returns the created Project. It rolls back on any failure, including a failed commit, and returns the commit error.
  - `DeleteProjectWithMembers(project_id)`: in one transaction, deletes the Project's members, then the Project. It checks the begin error before deferring the rollback, and returns the commit error.
  - `FindOneById(params)`: wraps `FindOneProjectById`.
  - `FindOneByIdAndRole(params)`: wraps `FindOneProjectByIdAndRole`.
  - `FindForUser(user_id)`: wraps `FindProjectsForUser`.
  - `UpdateOne(params)`: wraps `UpdateOneProject`.
- The exact parameter shape of `CreateProjectWithMember` is left to the implementer, but it must carry the org ID, the Project name, the member's user ID and the member role. It must not take a transaction or any pgx type.
- Transactions are owned by the repository through atomic, per-operation methods. Transactions are not exposed to the service: no `BeginTx`, no tx parameters and no `WithTx` closure. This was chosen because it is the simplest to stub and keeps commit and rollback bookkeeping in one place. The known cost is that these methods cannot share a transaction with another domain's repository. If deleting an organization later has to cascade to its Projects atomically, that will need a separate decision.

**Modified module: project service**

- `NewService` takes the context, the logger, a `ProjectRepository` and the `user.Service`, instead of the pool and `Queries`. The `Service` interface (the public methods) is unchanged, so the handlers, the router and the application domain's project service stub need no changes.
- `Create` looks up the user through the user service as before, then calls `CreateProjectWithMember` with role `owner`. The duplicate-name error from the database is still returned as is, so the handler's "duplicate key" check keeps working.
- `DeleteOne` calls `DeleteProjectWithMembers`. The exported-by-accident helper that deleted members inside a transaction it was handed is removed, since it isn't part of the `Service` interface and its job moves into the repository.
- `FindById`, `FindByIdAndRole`, `ListMyProjects` and `UpdateOne` call the matching repository method and keep their current error translation and return values.
- `main.go` builds the project repository from the context, the pool and `Queries`, and passes it to `project.NewService`, the same way the application repository is wired today.
- Code follows the repo conventions: a docstring with `@params`, `@return` and a description on every function and type, block-form control flow, no multi-line conditions, and snake_case locals and params.

## Testing Decisions

- A good test drives the service only through its `Service` interface and checks two things: what the service returns (values and errors), and what it did to the stubs (call counts and call args). It must not depend on how the service is put together internally.
- **Seam:** one domain logic test seam. The real service is built with `NewService(ctx, logger, stub_repo, stub_user_service)`. There is no HTTP and no database in these tests.
- New stubs in the project package's `stubs.go`:
  - `StubProjectRepository`: a `*_return` / `*_err` field pair, a call counter and recorded call args for every repository method, plus a `Clear()`.
  - `StubUserService`: configurable `FindById` return and error.
- Service tests go next to the service and use table-driven `t.Run` cases with `got_*`/`want_*` naming. Cases:
  - **Create:** the user lookup errors, so the repository is never called. The user is found, so `CreateProjectWithMember` is called once with the org ID, the name, the user's ID and role `owner`. A repository error, including a duplicate-key error, is returned unchanged.
  - **DeleteOne:** calls `DeleteProjectWithMembers` once with the Project ID and propagates its error.
  - **FindById and FindByIdAndRole:** a "no rows" error returns nil with no error. Any other error returns the generic error as today. On success the row is returned.
  - **ListMyProjects:** "no rows" returns an empty slice. Any other error returns an error. On success the rows are returned.
  - **UpdateOne:** sends the name and Project ID from the input and sets an `UpdatedAt`. It propagates the repository's error.
- The repository is not unit-tested. It is a thin wrapper over sqlc and the tests have no database, which matches the application and deployment repositories. Its transaction behaviour is checked by code review against the two bugs it replaces.
- The existing `tests/*_integration_test.go` HTTP tests stub `project.Service` and must pass unchanged. That is the regression check for the API contract.
- Prior art: the application domain's service tests and stubs, and the deployment domain's repository stubs.

## Out of Scope

- Any change to the project API's behaviour, routes, handlers or DTOs.
- Checking that the creator belongs to the target Organization when creating a Project (a known authorization gap).
- Making `FindByIdAndRole` actually filter by the roles it is given (a known gap: the `roles` argument is ignored).
- The deletion rules for a Project that still has Applications. Today the foreign key makes this a 500, and the block, cascade or soft-delete decision is still open.
- Making Project names unique per Organization rather than globally.
- An organization repository. The organization service has the same pool-plus-`Queries` shape and the same transaction bugs, and should get the same treatment in a follow-up.
- New project endpoints (per-org listing, project members).
- Composable or cross-repository transactions.

## Further Notes

- This came out of the review of unimplemented management endpoints for the organization, project and application domains. A repository-backed project service is a prerequisite for unit-testing the upcoming project endpoints.
- Transaction options considered: atomic methods on the repository (chosen); the repository exposing a begin-transaction call with every method accepting the transaction (rejected because it puts pgx types and commit/rollback bookkeeping back in the service and makes stubs heavy); and a `WithTx(fn)` closure (rejected as more flexibility than is needed right now).
- There is no `CONTEXT.md` yet. "Project", "Organization" and "project member" follow the backend's domain packages and tables.
