-- name: CreateApplicationDeployment :one
INSERT INTO "application_deployments" (
  app_id,
  artifacts_path,
  build_path,
  variables_snapshot_json,
  version_number,
  status,
  storage_service
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING *;

-- name: FindCurrentDeployment :one
SELECT 
  * 
FROM 
  "application_deployments" AS AD
WHERE 
  AD.app_id = $1;

-- name: CreateDeploymentInstance :one
INSERT INTO "deployment_instances" (
  app_id,
  deployment_id,
  container_name,
  host_port,
  container_port
)
VALUES ($1, $2, $3, $4, $5)
RETURNING *;

-- name: UpdateDeploymentStatus :one
UPDATE "application_deployments"
  SET status = $2
WHERE
  app_dp_id = $1
RETURNING 
  app_dp_id, status;

-- name: UpdateDeploymentInstanceStatus :one
UPDATE "deployment_instances"
  SET status = $2
WHERE
  instance_id = $1
RETURNING 
  instance_id, status;