# Keep Deployment Instance status in sync with the container

Status: needs-triage

Spec: `../list-project-applications-spec.md`

## Problem

Application `status` (issue 03) is read from the Application's newest `deployment_instances` row. That row doesn't reflect the container's real state:

1. **`RUNNING` is never written.** The Start step creates the instance as `DEPLOY_INSTANCE_STATUS_STOPPED`. `masbro-worker`'s `Start` runs the container and sets `HostPort` on the instance, but returns its `Status` unchanged. The deployment service then saves that `STOPPED` value. Nothing in the codebase writes `DEPLOY_INSTANCE_STATUS_RUNNING`, so every deployed Application shows `stopped`.
2. **A dead container goes unnoticed.** Nothing updates an instance when its container exits, crashes or is removed, so a dead container would still show `running` once (1) is fixed.
3. **Older instances are never retired.** Each Deployment's Start step creates a new instance and leaves the previous ones as they are. Issue 03 only reads the newest instance, so this doesn't break the dashboard, but older rows keep claiming whatever state they last had.

## What to build

- [ ] After `docker.Run` succeeds, the Start step saves the instance as `DEPLOY_INSTANCE_STATUS_RUNNING`. If the container fails to start, the instance stays `STOPPED`. This part is fully specified and can be done on its own.
- [ ] Container exits are reflected in `deployment_instances` (see the decision below), using the existing `UpdateDeploymentInstanceStatus` query.
- [ ] When a new instance becomes `RUNNING`, the Application's older instances are marked `STOPPED`, or whatever the decision on (3) is.
- [ ] Domain logic tests at the deployment service seam (Start sets `RUNNING`, a failed Run leaves `STOPPED`) and at the reconciler's seam, using the existing docker and repository stubs.

## Decisions needed

1. **How container exits are detected.**
   - **A. Docker events:** a listener goroutine subscribes to container `die`/`stop`/`start` events through the shared docker client, started from `main.go` like `deployment.Listener`. Updates are near-instant, but events are missed while the backend is down, so it also needs (B) on startup.
   - **B. Periodic reconcile:** every N seconds, inspect the containers of every `RUNNING` instance and update any that changed. Simple and self-healing, but status lags by up to N.
   - **C. Inspect on read:** issue 03's list call inspects containers live. Always accurate, but it adds Docker calls to a dashboard read and couples the application domain to Docker.

   Recommendation: A, plus a one-off B run at startup.
2. **Old containers.** When a new Deployment starts, should the previous container be stopped (a rolling replace), or keep running alongside it? This decides whether (3) only updates rows or also stops containers. Today the pipeline leaves the old container running.
3. **More instance states.** Does the dashboard need `starting`, `exited` or `crashed` (with exit code), or are `RUNNING`/`STOPPED` enough? Adding states extends `DEPLOY_INSTANCE_STATUS_*` and the status table in issue 03.

## Notes

- The API contract of issue 03 doesn't change. Once this lands, its `status` becomes accurate.
- `host_id` exists on `deployment_instances` but there is no hosts table yet (issue 10). Reconciliation assumes the single local Docker host for now.
