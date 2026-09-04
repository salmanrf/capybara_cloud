ALTER TABLE "application_configs"
  ADD COLUMN "runtime_type" text NOT NULL DEFAULT 'DOCKER_CONTAINER'::text;