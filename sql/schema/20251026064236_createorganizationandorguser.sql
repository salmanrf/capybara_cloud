-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "organizations" (
  "org_id" uuid PRIMARY KEY DEFAULT uuid_generate_v4(),
  "name" varchar(250) UNIQUE NOT NULL,
  "created_at" timestamp DEFAULT now(),
  "updated_at" timestamp DEFAULT now()
);

CREATE TABLE IF NOT EXISTS "organization_users" (
  "org_id" uuid NOT NULL,
  "user_id" uuid NOT NULL,
  "role" varchar(25) NOT NULL,
  "created_at" timestamp DEFAULT now(),
  PRIMARY KEY("org_id", "user_id")
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "organizations";
DROP TABLE "organization_users";
-- +goose StatementEnd
