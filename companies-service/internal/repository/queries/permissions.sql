-- name: GiveAccess :exec
INSERT INTO company.permissions (company_uuid,
                                 account_uuid,
                                 domain,
                                 mask,
                                 change_by)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUserPermissionMask :one
SELECT mask, is_active
FROM company.permissions
WHERE company_uuid = $1
  AND account_uuid = $2
  AND domain = $3
  AND is_active = true;

-- name: UpdatePermissions :exec
INSERT INTO company.permissions (
    company_uuid,
    account_uuid,
    domain,
    mask,
    change_by
) VALUES ($1, $2, $3, $4, $5)
ON CONFLICT (company_uuid, account_uuid, domain)
DO UPDATE SET
    mask = EXCLUDED.mask,
    change_by = EXCLUDED.change_by,
    updated_at = NOW();

-- name: GetUserAllPermissions :many
SELECT
    domain,
    mask
FROM company.permissions
WHERE company_uuid = $1 AND
      account_uuid = $2;
