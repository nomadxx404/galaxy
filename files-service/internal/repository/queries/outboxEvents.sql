-- name: CreateOutboxEvent :exec
INSERT INTO file.outbox_events (message_uuid,
                                   account_uuid,
                                   request_id,
                                   event_type,
                                   payload)
VALUES ($1, $2, $3, $4, $5);

-- name: GetUnprocessedEvents :many
SELECT message_uuid,
       account_uuid,
       event_type,
       payload,
       status,
       retry_count,
       last_error,
       created_at,
       processed_at,
       locked_until
FROM file.outbox_events
WHERE status = 'PENDING'
ORDER BY created_at ASC
LIMIT 50;

-- name: MarkEventProcessed :exec
UPDATE file.outbox_events
SET status       = 'PROCESSED',
    processed_at = NOW()
WHERE message_uuid = $1;

-- name: MarkEventFailed :exec
UPDATE file.outbox_events
SET retry_count = retry_count + 1,
    last_error  = $2
WHERE message_uuid = $1;
