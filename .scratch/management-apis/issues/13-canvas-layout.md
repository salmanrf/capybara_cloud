# Persist Applications canvas positions

Status: needs-triage

Spec: `../spec.md`

## What the design needs

App cards on the canvas can be dragged (`cursor: grab`), and the canvas has zoom controls.

## Decision needed

- Should positions be per Project (shared by all viewers) or per user?
- Shared: add `canvas_x` / `canvas_y` to `applications` and a `PATCH /api/projects/{project_id}/layout` that saves every position in one call.
- Per user: keep positions in the browser's `localStorage` with no backend change. That's enough until real-time collaboration matters.

Recommendation: start with `localStorage`. It's cheap and easy to undo.
