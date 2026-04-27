-- name: IsExistsRole :one
SELECT count(*) != 0
FROM company.roles
WHERE company_uuid = $1
  AND name = $2;

-- name: CreateRole :one
INSERT INTO company.roles (name, description, color, company_uuid)
VALUES ($1, $2, $3, $4)
RETURNING
    role_id,
    name,
    description,
    color,
    created_at,
    is_active;

-- name: GetRoles :many
SELECT role_id,
       name,
       color,
       is_active
FROM company.roles
WHERE company_uuid = $1;

-- name: GetRoleByUuid :one
SELECT role_id,
       name,
       description,
       color,
       created_at,
       is_active
FROM company.roles
WHERE company_uuid = $1
  AND role_id = $2;


-- name: UpdateRole :one
UPDATE company.roles
SET name        = COALESCE(sqlc.narg('name'), name),
    description = COALESCE(sqlc.narg('description'), description),
    color       = COALESCE(sqlc.narg('color'), color),
    updated_at  = NOW()
WHERE company_uuid = $1
  AND role_id = $2
RETURNING
    role_id,
    name,
    description,
    color,
    created_at,
    is_active;

-- name: DeleteRole :exec
UPDATE company.roles
SET is_active  = false,
    updated_at = NOW()
WHERE company_uuid = $1
  AND role_id = $2
  AND is_active = true;

-- name: GetRoleIdByName :one
SELECT role_id
FROM company.roles
WHERE company_uuid = $1
  AND name = $2;

-- name: DeleteAllRoles :many
UPDATE company.roles
SET is_active  = FALSE,
    updated_at = now()
WHERE company_uuid = $1
RETURNING role_id;
