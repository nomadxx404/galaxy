import hashlib
import json
import logging
from src.repository.logs_repository import LogsRepository
from src.schemas.logs_schemas import EventEnvelope

logger = logging.getLogger("uvicorn.error")


class LogsService:
    def __init__(self, repo: LogsRepository):
        self.repo = repo

    async def process_event(self, raw_event: dict) -> None:
        try:
            envelope = EventEnvelope.model_validate(raw_event)
            meta = envelope.metadata
            payload = envelope.payload

            payload_str = json.dumps(
                payload, sort_keys=True, ensure_ascii=False)
            body_hash = hashlib.sha256(payload_str.encode()).hexdigest()
            body_size = len(payload_str)

            await self.repo.insert_log(
                service=meta.service,
                method=meta.method,
                path=meta.path,
                status=meta.status or 200,
                account_uuid=meta.account_uuid,
                request_id=meta.request_id,
                body_hash=body_hash,
                body_size=body_size
            )

            logger.debug(f"Saved log from {meta.service}: {meta.path}")

        except Exception as e:
            logger.error(
                f"Error processing logging event: {e} | Raw: {raw_event}")
