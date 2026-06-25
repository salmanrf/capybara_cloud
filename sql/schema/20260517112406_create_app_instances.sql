-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "deployment_instances" (
  "instance_id" uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  "app_id" uuid NOT NULL,
  "deployment_id" uuid NOT NULL,
  "container_name" varchar(500) NOT NULL,
  "status" INT NOT NULL DEFAULT 0,
  "host_id" uuid NULL,
  "host_port" int NOT NULL,
  "container_port" int NOT NULL,
  "created_at" timestamp DEFAULT NOW(),
  "updated_at" timestamp DEFAULT NOW(),
  FOREIGN KEY(app_id) REFERENCES "applications"(app_id),
  FOREIGN KEY(deployment_id) REFERENCES "application_deployments"(app_dp_id)
);
-- +goose StatementEnd

ALTER TABLE "application_deployments"
ADD COLUMN "status" INT NOT NULL DEFAULT 1,
ADD COLUMN "build_path" text NOT NULL,
DROP COLUMN "container_name",
DROP COLUMN "process_name";

-- +goose Down
-- +goose StatementBegin
DROP TABLE "deployment_instances";
-- +goose StatementEnd
