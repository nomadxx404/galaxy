using auth_service.DTO.Entity;
using auth_service.Repositories.Interfaces;
using Dapper;
using Npgsql;

namespace auth_service.Repositories
{
    public class OutboxEventsRepository : IOutboxEventsRepository
    {
        public async Task<IEnumerable<OutboxEventEntity>> GetUnprocessedEvents(NpgsqlConnection conn)
        {
            return await conn.QueryAsync<OutboxEventEntity>(
               """
               SELECT message_uuid as MessageUuid,
                      account_uuid as AccountUuid,
                      event_type   as EventType,
                      payload      as Payload,
                      status       as Status,
                      retry_count  as RetryCount,
                      last_error   as LastError,
                      created_at   as CreatedAt,
                      processed_at as ProcessedAt,
                      locked_until as LockedUntil
               FROM auth.outbox_events
               WHERE status = 'PENDING'
               ORDER BY created_at ASC
               LIMIT 50
               """);
        }

        public async Task MarkEventFailed(string message_uuid, string last_error, NpgsqlConnection conn, NpgsqlTransaction? transaction = null)
        {
            await conn.ExecuteAsync(
                """
                UPDATE auth.outbox_events
                SET retry_count = retry_count + 1,
                    last_error  = @last_error
                WHERE message_uuid = @message_uuid;
                """,
                new
                {
                    message_uuid = message_uuid,
                    last_error = last_error,
                }, transaction);
        }

        public async Task MarkEventProcessed(string message_uuid, NpgsqlConnection conn, NpgsqlTransaction? transaction = null)
        {
            await conn.ExecuteAsync(
                """
                UPDATE auth.outbox_events
                SET status       = 'PROCESSED',
                    processed_at = NOW()
                WHERE message_uuid = @message_uuid;
                """,
                new
                {
                    message_uuid = message_uuid,
                }, transaction);
        }

        public async Task OutboxEventsCreate(string message_uuid, string account_uuid, string request_id, string eventType, object payload, NpgsqlConnection conn, NpgsqlTransaction? transaction = null)
        {
            await conn.ExecuteAsync(
                """
                INSERT INTO auth.outbox_events (message_uuid,
                                                account_uuid,
                                                request_id,
                                                event_type,
                                                payload)
                VALUES (@message_uuid, 
                        @account_uuid, 
                        @request_id, 
                        @eventType, 
                        @payload::jsonb)
                """,
                new
                {
                    message_uuid = message_uuid,
                    account_uuid = account_uuid,
                    request_id = request_id,
                    eventType = eventType,
                    payload = payload,
                }, transaction);
        }
    }
}
