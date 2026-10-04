# Node status and capacity

Status: needs-triage

Spec: `../spec.md`

## What the design needs

The sidebar footer shows a health dot and "node-01 · 3 of 8 GB used".

## Decisions needed

1. **Hosts model.** `deployment_instances.host_id` exists, but there is no hosts table. Is a single-node install the only target for now (report the local Docker host), or do we model hosts?
2. **Metric source.** Total memory and "used" (host memory, or the sum of container limits?) from the Docker daemon (`docker info`, container stats) through the shared docker client.
3. **Scope.** Should it be global (`GET /api/nodes`) or per Organization? Who is allowed to see it?
