import asyncpg
import logging
from typing import Optional
from src.config.config import settings

logger = logging.getLogger("uvicorn.error")


class PostgresManager:
    def __init__(self):
        self._pool: Optional[asyncpg.Pool] = None

    async def connect(self) -> None:
        if self._pool is None:
            try:
                self._pool = await asyncpg.create_pool(
                    host=settings.POSTGRES_HOST,
                    port=settings.POSTGRES_PORT,
                    user=settings.POSTGRES_USER,
                    password=settings.POSTGRES_PASSWORD,
                    database=settings.POSTGRES_DB,
                    min_size=5,
                    max_size=20,
                    command_timeout=60,
                )
                logger.info(
                    "Successfully connected to PostgreSQL (galaxy_logs)")
            except Exception as e:
                logger.error(f"Failed to connect to PostgreSQL: {e}")
                raise e

    async def disconnect(self) -> None:
        if self._pool:
            await self._pool.close()
            self._pool = None
            logger.info("PostgreSQL connection pool closed")

    @property
    def pool(self) -> asyncpg.Pool:
        """Свойство для получения объекта пула в репозиториях"""
        if self._pool is None:
            raise RuntimeError(
                "PostgresManager is not connected. Call connect() first.")
        return self._pool


db_manager = PostgresManager()
