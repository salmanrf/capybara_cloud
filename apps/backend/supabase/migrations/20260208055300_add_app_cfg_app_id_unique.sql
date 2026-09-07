CREATE UNIQUE INDEX CONCURRENTLY app_cfg_app_id
ON application_configs (app_id);

ALTER TABLE application_configs
ADD CONSTRAINT unique_app_cfg_app_id
UNIQUE USING INDEX app_cfg_app_id;