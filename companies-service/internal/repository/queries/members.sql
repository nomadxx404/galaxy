-- name: CreateMember :exec
INSERT INTO company.members (company_uuid,
                             account_uuid,
                             role_id,
                             is_owner)
VALUES ($1, $2, $3, $4);

-- name: UpdateRoleMember :execrows
UPDATE company.members
SET role_id = $3
WHERE company_uuid = $1
  AND account_uuid = $2;

-- name: SetOwner :execrows
UPDATE company.members
SET is_owner   = TRUE,
    updated_at = NOW()
WHERE company_uuid = $1
  AND account_uuid = $2;


-- name: DeleteMember :exec
UPDATE company.members
SET is_active  = FALSE,
    updated_at = NOW()
WHERE company_uuid = $1
  AND account_uuid = $2;


-- name: GetMembersStatuses :many
SELECT account_uuid, is_owner
FROM company.members
WHERE company_uuid = $1
  AND account_uuid IN ($2, $3);

-- name: GetMembers :many
SELECT m.account_uuid,
       m.role_id,
       m.created_at,
       r.name  AS role_name,
       r.color AS role_color
FROM company.members m
         JOIN company.roles r ON m.role_id = r.role_id
WHERE m.company_uuid = $1
ORDER BY m.created_at DESC;

