from pydantic import BaseModel
from typing import Any, Optional


class EventMetadata(BaseModel):
    service: str
    request_id: str
    account_uuid: str
    method: str
    path: str
    status: Optional[int] = 0
    body_hash: Optional[str] = ""
    body_size: Optional[int] = 0


class EventEnvelope(BaseModel):
    event_type: str
    metadata: EventMetadata
    payload: Any
