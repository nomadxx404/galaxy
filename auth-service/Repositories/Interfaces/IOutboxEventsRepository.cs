using auth_service.DTO.Entity;
using Npgsql;

namespace auth_service.Repositories.Interfaces
{
    public interface IOutboxEventsRepository
    {
        public Task OutboxEventsCreate(string message_uuid, string account_uuid, string request_id, string eventType, object payload, NpgsqlConnection conn, NpgsqlTransaction? transaction = null);
        public Task<IEnumerable<OutboxEventEntity>> GetUnprocessedEvents(NpgsqlConnection conn);
        public Task MarkEventProcessed(string message_uuid, NpgsqlConnection conn, NpgsqlTransaction? transaction = null);
        public Task MarkEventFailed(string message_uuid, string last_error, NpgsqlConnection conn, NpgsqlTransaction? transaction = null);
    }
}
