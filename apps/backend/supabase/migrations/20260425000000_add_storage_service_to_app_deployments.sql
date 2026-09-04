ALTER TABLE "application_deployments"
ADD COLUMN "storage_service" varchar(50) NOT NULL DEFAULT 'localfs';

ALTER TABLE "application_configs"
ADD COLUMN "port" integer NOT NULL DEFAULT 80;

ALTER TABLE "application_deployments"
ADD COLUMN "version_number" integer NOT NULL DEFAULT 1;

CREATE UNIQUE INDEX idx_app_deployments_app_id_version_number 
ON "application_deployments"("app_id", "version_number");


