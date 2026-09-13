ALTER TABLE "application_deployments"
  DROP COLUMN "container_img_name",
  DROP COLUMN "container_registry",
  ADD COLUMN "container_registry" text,
  ADD COLUMN "container_namespace" text,
  ADD COLUMN "container_repository" text,
  ADD COLUMN "container_tag" text;
