using System.Text.Json.Serialization;

namespace auth_service.DTO.Entity
{
    public record OutboxEventEntity(
        string MessageUuid,
        string AccountUuid,
        string EventType,
        string Payload,
        string Status,
        int RetryCount,
        string? LastError,
        DateTime CreatedAt,
        DateTime? ProcessedAt,
        DateTime? LockedUntil
    );
}

public record EventMetadata(
    [property: JsonPropertyName("service")] string Service,
    [property: JsonPropertyName("request_id")] string RequestId,
    [property: JsonPropertyName("account_uuid")] string AccountUuid,
    [property: JsonPropertyName("method")] string Method,
    [property: JsonPropertyName("path")] string Path
);

public record EventEnvelope(
    [property: JsonPropertyName("event_type")] string EventType,
    [property: JsonPropertyName("metadata")] EventMetadata Metadata,
    [property: JsonPropertyName("payload")] object Payload
);
