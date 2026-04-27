using auth_service.Config;
using auth_service.Repositories.Interfaces;
using Npgsql;
using System.Text.Json;

namespace auth_service.Infrastructure.Kafka
{
    public interface IOutboxEmitter
    {
        public Task EmitOutbox(string eventType, object payload, NpgsqlConnection conn, NpgsqlTransaction transaction, string? account_uuid = null);
    }

    public class OutboxEmitter: IOutboxEmitter
    {
        private readonly IUserContext _userContext;
        private readonly IOutboxEventsRepository _outboxEventsRepository;
        private readonly AppConfig _appConfig;

        public OutboxEmitter(
            IUserContext userContext,
            IOutboxEventsRepository outboxEventsRepository,
            AppConfig appConfig)
        {
            _userContext = userContext;
            _outboxEventsRepository = outboxEventsRepository;
            _appConfig = appConfig;
        }

        public async Task EmitOutbox(
            string eventType,
            object payload,
            NpgsqlConnection conn,
            NpgsqlTransaction transaction,
            string? account_uuid = null)
        {
            var envelope = new EventEnvelope(
                EventType: eventType,
                Metadata: new EventMetadata(
                    Service: _appConfig.KafkaConfig.EVENTS_TOPIC,
                    RequestId: _userContext.RequestId,
                    AccountUuid: account_uuid ?? _userContext.Account_uuid ?? string.Empty,
                    Method: _userContext.Method,
                    Path: _userContext.Path
                ),
                Payload: payload
            );

            var payloadJson = JsonSerializer.Serialize(envelope);

            await _outboxEventsRepository.OutboxEventsCreate(
                message_uuid: Guid.CreateVersion7().ToString(),
                account_uuid: envelope.Metadata.AccountUuid,
                request_id: envelope.Metadata.RequestId,
                eventType: eventType,
                payload: payloadJson,
                conn: conn,
                transaction: transaction);
        }
    }
}
