# Application public domain and runtime label

Status: needs-triage

Spec: `../spec.md`

## What the design needs

Each app card shows either a public domain link (`storefront.masbro.app ↗`) or "No public domain", and a runtime tag (`node 20`).

## Decisions needed

1. **Domains.** Is there a reverse proxy or ingress that routes hostnames to containers? Are domains generated (`<app>.<base-domain>`), custom, or both? Where are they stored: a column on `applications`, or a `domains` table for several per app? Is a domain opt-in, given the design's "No public domain" state?
2. **Runtime label.** `application_configs.runtime_type` is `DOCKER_CONTAINER`, not a language version. Should the build step detect the runtime (e.g. from `package.json` engines or the Dockerfile template) and record it on the Deployment, or should the user set it in the config?
