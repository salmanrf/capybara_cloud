ALTER TABLE "application_deployments"
ADD COLUMN "container_img_name" text NOT NULL,
ADD COLUMN "container_registry" text NOT NULL;
