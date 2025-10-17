-- +goose Up
-- +goose StatementBegin
CREATE TABLE IF NOT EXISTS "users" (
  "user_id" uuid PRIMARY KEY DEFAULT 'gen_uuid_v4()',
  "username" varchar(100) UNIQUE NOT NULL,
  "role" varchar,
  "created_at" timestamp DEFAULT 'now()',
  "updated_at" timestamp DEFAULT 'now()'
);
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
DROP TABLE "users";
-- +goose StatementEnd
