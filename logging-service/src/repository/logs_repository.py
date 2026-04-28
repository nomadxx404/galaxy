import asyncpg
from typing import Optional


class LogsRepository:
    def __init__(self, pool: asyncpg.Pool):
        self.pool = pool

    async def insert_log(
        self,
        service: str,
        method: str,
        path: str,
        status: int,
        account_uuid: str,
        request_id: str,
        body_hash: str,
        body_size: int
    ) -> None:
        await self.pool.execute(
            """
            INSERT INTO log.logs (service, method, path, status,
                                account_uuid, request_id, body_hash, body_size)
            VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
            """,
            service,
            method,
            path,
            status,
            account_uuid,
            request_id,
            body_hash,
            body_size
        )
