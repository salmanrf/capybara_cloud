# Project card summary on the org Overview

Status: needs-triage
Blocked by: 01, 03, 04

Spec: `../spec.md`

## What the design needs

Each project card shows: the Application chips with status dots, "N apps", the latest deploy outcome and relative time across the Project ("deployed 12m ago" / "failed 1h ago"), the repo count (11) and the environment badge (09).

## Decision needed

How the frontend gets this data:

- **A. Compose on the client:** 01, then 03 (`GET /applications?project_id=`) for each Project. This adds no new endpoint, but it's N+1 requests.
- **B. Embed in 01:** add `applications: [{ app_id, name, status }]` and `last_deployment: { outcome, updated_at }` to each entry of `GET /projects?org_id={org_id}`. One request, but a heavier query.
- **C. Separate endpoint:** `GET /projects/summary?org_id={org_id}`.

Recommendation: B, reusing the `status` and `outcome` derivations from 03.
