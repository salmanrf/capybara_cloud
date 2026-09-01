-- +goose Up
-- +goose StatementBegin
ALTER TABLE "users"
ADD COLUMN "email" VARCHAR(100) UNIQUE NOT NULL,
ADD COLUMN "full_name" VARCHAR(500) NOT NULL,
ADD COLUMN "hashed_password" TEXT NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "users"
DROP COLUMN "email",
DROP COLUMN "full_name";
-- +goose StatementEnd
