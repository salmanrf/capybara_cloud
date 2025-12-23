-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "applications" (
  "app_id" uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  "project_id" uuid NOT NULL,
  "type" varchar(50) NOT NULL,
  "name" varchar(100) NOT NULL,
  "created_at" timestamp DEFAULT NOW(),
  "updated_at" timestamp DEFAULT NOW(),
  FOREIGN KEY(project_id) REFERENCES "projects"(project_id)
);

CREATE TABLE IF NOT EXISTS "application_deployments" (
  "app_dp_id" uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  "app_id" uuid NOT NULL,
  "artifacts_path" text NOT NULL,
  "process_name" varchar(255) NOT NULL,
  "container_name" varchar(255) NOT NULL,
  "variables_snapshot_json" jsonb,
  "created_at" timestamp DEFAULT NOW(),
  "updated_at" timestamp DEFAULT NOW(),
  FOREIGN KEY(app_id) REFERENCES "applications"(app_id)
);

CREATE TABLE IF NOT EXISTS "application_configs" (
  "app_cfg_id" uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  "app_id" uuid NOT NULL,
  "variables_json" jsonb,
  "created_at" timestamp DEFAULT NOW(),
  "updated_at" timestamp DEFAULT NOW(),
  FOREIGN KEY(app_id) REFERENCES "applications"(app_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "application_deployments";
DROP TABLE "application_configs";
DROP TABLE "applications";
-- +goose StatementEnd
