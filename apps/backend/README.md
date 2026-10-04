# Backend

Go API server (chi + pgx) for Capybara Cloud. Commands below run from the repository root with
`nx`; see the root `README.md` for installation.

## Database & Migration

Local database runs on Supabase CLI. Migrations live in `apps/backend/supabase/migrations`.

```
nx supabase-start backend  # supabase start
nx supabase-reset backend  # supabase db reset, applies all migrations
```

## Environment Variables

Loaded from `apps/backend/.env`.

```
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
DOCKER_USER=
DOCKER_ACCESS_TOKEN=
BASE_ARTIFACT_PATH=
BASE_BUILD_PATH=
DOCKER_TEMPLATES_DIR=
```
