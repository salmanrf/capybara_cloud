-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "projects" (
  "project_id" uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  "org_id" uuid NOT NULL,
  "name" varchar(250) UNIQUE NOT NULL,
  "created_at" timestamp DEFAULT now(),
  "updated_at" timestamp DEFAULT now(),
  FOREIGN KEY(org_id) REFERENCES "organizations"(org_id)
);

CREATE TABLE IF NOT EXISTS "project_members" (
  "project_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "role" varchar(25) NOT NULL,
  "created_at" timestamp DEFAULT now(),
  PRIMARY KEY(project_id, user_id),
  FOREIGN KEY(project_id) REFERENCES "projects"(project_id),
  FOREIGN KEY(user_id) REFERENCES "users"(user_id)
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "project_members";
DROP TABLE "projects";
-- +goose StatementEnd
