-- +goose Up
-- +goose NO TRANSACTION
CREATE UNIQUE INDEX CONCURRENTLY app_cfg_app_id
ON application_configs (app_id);

ALTER TABLE application_configs
ADD CONSTRAINT unique_app_cfg_app_id
UNIQUE USING INDEX app_cfg_app_id;

-- +goose Down
-- +goose NO TRANSACTION
ALTER TABLE application_configs
DROP CONSTRAINT IF EXISTS unique_app_cfg_app_id;

DROP INDEX CONCURRENTLY IF EXISTS app_cfg_app_id;