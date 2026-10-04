# Access links between Applications

Status: needs-triage

Spec: `../spec.md`

## What the design needs

On the Applications canvas you drag from one app's port to another to create an access link. The edge is labelled with a variable name (`API_URL`). The link panel shows `storefront-web → api-gateway` and explains that "the endpoint is injected at deploy time" as `API_URL=http://api-gateway.internal:3000`, with "Rename variable" and "Remove link" buttons.

Likely routes: `GET /api/projects/{project_id}/links`, `POST` (create), `PATCH /api/links/{link_id}` (rename the variable), `DELETE /api/links/{link_id}`.

## Decisions needed

1. **Data model.** A new `application_links` table (`source_app_id`, `target_app_id`, `variable_name`, unique per source + variable). Both apps must be in the same Project (and the same environment, see 09).
2. **Private networking.** `*.internal` hostnames need the worker to put a Project's containers on a shared Docker network with DNS aliases. Today containers are started individually with host port mapping (`packages/shared-go/docker`). This is the biggest piece of work here.
3. **Injection timing.** Should the variable be merged into the env at deploy time (snapshotted in `variables_snapshot_json`)? Does changing a link redeploy the source app, or only apply on the next deploy?
4. **Conflicts.** What happens when a link's variable name collides with a key in the app's own config variables? Which one wins?
5. **Target port.** Is it the target's `application_configs.port`?
