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
  container_port,
  status
)
VALUES ($1, $2, $3, $4, $5, $6)
RETURNING *;

-- name: UpdateDeploymentStatus :one
UPDATE "application_deployments"
SET 
  status = $2,
  updated_at = NOW()
WHERE
  app_dp_id = $1
RETURNING 
  *;

-- name: UpdateDeploymentInstanceStatus :one
UPDATE "deployment_instances"
  SET status = $2
WHERE
  instance_id = $1
RETURNING 
  instance_id, status;

-- name: UpdateDeployment :one
UPDATE "application_deployments"
SET
  app_id = $2,
  artifacts_path = $3,
  variables_snapshot_json = $4,
  storage_service = $5,
  version_number = $6,
  status = $7,
  build_path = $8,
  container_img_name = $9,
  container_registry = $10,
  build_path = $11,
  updated_at = NOW()
WHERE app_dp_id = $1
RETURNING *;

-- name: UpdateDeploymentInstance :one
INSERT INTO "deployment_instances" (
  instance_id,
  app_id,
  deployment_id,
  container_name,
  host_port,
  container_port,
  status
)
VALUES ($1, $2, $3, $4, $5, $6, $7)
ON CONFLICT (instance_id)
DO UPDATE SET 
  app_id = $2,
  deployment_id = $3,
  container_name = $4,
  host_port = $5,
  container_port = $6,
  updated_at = NOW()
RETURNING *;