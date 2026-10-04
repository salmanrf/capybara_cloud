# Environments (production / staging)

Status: needs-triage

Spec: `../spec.md`

## What the design needs

Each Project card on the Overview has an environment badge ("production", "staging"). The project header has an environment switcher ("● production ▾"), and the project sidebar card says "Project · production".

## Decision needed

- **A. A label on the Project:** an `environment` column on `projects`. Cheap, but then the switcher has nothing to switch between.
- **B. A real environment entity:** each Project has environments, and Applications, configs, Deployments and access links are scoped per environment. The switcher needs this. It touches almost every table and the deploy pipeline.

Pick one before 07 and 08 are built, because B changes the shapes they return.
