package kafka

import (
	"context"
	"encoding/json"
	"files-service/internal/repository/db"
	"files-service/pkg/config"
	"files-service/pkg/response"
	usercontext "files-service/pkg/user_context"

	"github.com/google/uuid"
)

type EventEnvelope struct {
	EventType string        `json:"event_type"`
	Metadata  EventMetadata `json:"metadata"`
	Payload   any           `json:"payload"`
}

type EventMetadata struct {
	Service     string `json:"service"`
	RequestID   string `json:"request_id"`
	AccountUUID string `json:"account_uuid"`
	Method      string `json:"method"`
	Path        string `json:"path"`
}

func EmitOutbox(
	ctx context.Context,
	dbExecutor db.DBTX,
	eventType string,
	payload any) error {

	cfg := config.GetConfig()
	envelope := EventEnvelope{
		EventType: eventType,
		Metadata: EventMetadata{
			Service:     cfg.Kafka.EVENTS_TOPIC_FILE,
			RequestID:   usercontext.GetRequestId(ctx),
			AccountUUID: usercontext.GetAccountUuid(ctx),
			Method:      usercontext.GetMethod(ctx),
			Path:        usercontext.GetPath(ctx),
		},
		Payload: payload,
	}

	payloadBytes, err := json.Marshal(envelope)
	if err != nil {
		return err
	}

	q := db.New(dbExecutor)

	err = q.CreateOutboxEvent(ctx, db.CreateOutboxEventParams{
		MessageUuid: uuid.Must(uuid.NewV7()).String(),
		AccountUuid: envelope.Metadata.AccountUUID,
		RequestID:   envelope.Metadata.RequestID,
		EventType:   eventType,
		Payload:     payloadBytes,
	})

	if err != nil {
		return &response.ApiError{
			Status:  500,
			Message: "Ошибка записи события в Outbox",
			Data:    err,
		}
	}

	return nil
}
