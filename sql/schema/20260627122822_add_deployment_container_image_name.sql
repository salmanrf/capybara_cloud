-- +goose Up
-- +goose StatementBegin
ALTER TABLE "application_deployments"
ADD COLUMN "container_img_name" text NOT NULL,
ADD COLUMN "container_registry" text NOT NULL;
-- +goose StatementEnd

-- +goose Down
-- +goose StatementBegin
ALTER TABLE "application_deployments"
DROP COLUMN "container_img_name",
DROP COLUMN "container_registry";
-- +goose StatementEnd
