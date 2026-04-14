-- name: CreateCompany :one
INSERT INTO company.companies (company_uuid,
                               name,
                               description,
                               logo_file_id,
                               billing_status)
VALUES ($1, $2, $3, $4, $5)
RETURNING
    company_uuid,
    name,
    description,
    logo_file_id,
    billing_status;

-- name: GetCompanies :many
SELECT com.company_uuid,
       com.name,
       com.logo_file_id,
       com.billing_status,
       com.is_active
FROM company.companies com
         JOIN company.members mem ON com.company_uuid = mem.company_uuid
WHERE mem.account_uuid = $1;

-- name: UpdateCompany :one
UPDATE company.companies
SET name         = COALESCE(sqlc.narg('name'), name),
    description  = COALESCE(sqlc.narg('description'), description),
    logo_file_id = COALESCE(sqlc.narg('logo_file_id'), logo_file_id),
    updated_at   = NOW()
WHERE company_uuid = $1
RETURNING
    company_uuid,
    name,
    description,
    logo_file_id,
    billing_status;

-- name: GetCompanyByUuid :one
SELECT com.company_uuid,
       com.name,
       com.description,
       com.logo_file_id,
       com.billing_status,
       com.billing_until,
       com.created_at,
       com.is_active
FROM company.companies com
         JOIN company.members mem ON com.company_uuid = mem.company_uuid AND
                                     mem.account_uuid = $1
WHERE com.company_uuid = $2;

-- name: IsCompanyOwner :one
SELECT EXISTS (SELECT 1
               FROM company.members
               WHERE company_uuid = $1
                 AND account_uuid = $2
                 AND is_owner = TRUE
                 AND is_active = TRUE);

-- name: DeleteCompany :exec
UPDATE company.companies
SET is_active  = false,
    updated_at = NOW()
WHERE company_uuid = $1;
