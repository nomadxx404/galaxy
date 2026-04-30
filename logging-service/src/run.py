import asyncio
import logging
from contextlib import asynccontextmanager
from fastapi import FastAPI
import uvicorn

from src.config.config import settings
from src.data.postgres import db_manager
from src.repository.logs_repository import LogsRepository
from src.service.logs_service import LogsService
from src.kafka.consumer import LoggingConsumer

logging.basicConfig(level=logging.INFO)
logger = logging.getLogger("uvicorn.error")


@asynccontextmanager
async def lifespan(app: FastAPI):
    await db_manager.connect()

    repo = LogsRepository(db_manager.pool)
    service = LogsService(repo)

    consumer = LoggingConsumer(
        brokers=settings.BROKERS,
        topics=settings.TOPIC,
        logs_service=service
    )
    kafka_task = asyncio.create_task(consumer.start())

    logger.info("Logging Service is fully started and listening to Kafka")

    yield

    logger.info("Shutting down Logging Service...")

    await consumer.stop()
    kafka_task.cancel()

    await db_manager.disconnect()
    logger.info("Bye!")

app = FastAPI(
    title="Galaxy Logging Service",
    version="1.0.0",
    lifespan=lifespan
)


@app.get("/health")
async def health_check():

    return {
        "status": "healthy",
        "service": "logging-service",
        "database": "connected" if db_manager._pool else "disconnected"
    }

if __name__ == "__main__":
    uvicorn.run(
        "src.run:app",
        host="0.0.0.0",
        port=8000,
        reload=False,
    )
