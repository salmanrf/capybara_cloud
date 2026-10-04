# Git connections, repos and commit SHAs

Status: needs-triage

Spec: `../spec.md`

## What the design needs

- A "Git connections" item in the org sidebar nav (no screen designed yet).
- "N repos" in each project card footer.
- A short commit SHA (`8d02b1e`) on each app card.

## Decisions needed

Today deployments come from uploaded `.tar.gz` bundles only, so there is no repository or commit to show.

1. Which providers come first (GitHub App? generic deploy key?), and is the connection made at the Organization level?
2. Is a repo linked per Application (repo + branch + root dir)? "Repos per project" would then be the distinct repos across its apps.
3. Do Deployments record a `commit_sha`? For uploads, the CLI could send it as optional metadata even before git connections exist. That would let the card show a SHA early.
