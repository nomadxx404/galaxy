-- name: CreateFile :one
INSERT INTO file.files (entity_type,
                        entity_id,
                        name,
                        size,
                        content_type,
                        storage_key,
                        created_by)
VALUES ($1, $2, $3, $4, $5, $6, $7)
RETURNING file_id, created_at;

-- name: GetFiles :many
SELECT file_id,
       entity_type,
       entity_id,
       name,
       size,
       content_type,
       storage_key,
       status,
       created_by,
       created_at,
       is_active
FROM file.files
WHERE file_id = ANY($1::bigint[])
  AND entity_id = $2;

-- name: GetFileById :one
SELECT name,
       size,
       content_type,
       storage_key
FROM file.files
WHERE file_id = $1;

-- name: MarkFileAsProcessed :execrows
UPDATE file.files
SET status = 'PROCESSED',
    entity_id = $2,
    updated_at = now()
WHERE file_id = $1
  AND entity_id IS NULL;

-- name: DeleteFile :one
UPDATE file.files
SET is_active  = FALSE,
    status = 'DELETED',
    updated_at = now()
WHERE file_id = $1
  AND entity_id = $2
RETURNING storage_key;

-- name: DeleteAllFilesEntity :many
UPDATE file.files
SET is_active  = FALSE,
    status = 'DELETED',
    updated_at = now()
WHERE entity_id = $1
RETURNING storage_key;
