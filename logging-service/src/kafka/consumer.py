import json
import logging
from aiokafka import AIOKafkaConsumer
from src.service.logs_service import LogsService

logger = logging.getLogger("uvicorn.error")


class LoggingConsumer:
    def __init__(self, brokers: str, topics: list[str], logs_service: LogsService):
        self.consumer = AIOKafkaConsumer(
            *topics,
            bootstrap_servers=brokers,
            group_id="logging-service-group",
            enable_auto_commit=True,
            auto_commit_interval_ms=5000,
            value_deserializer=lambda m: json.loads(m.decode('utf-8')),
            auto_offset_reset="earliest"
        )
        self.logs_service = logs_service
        self._running = False

    async def start(self):
        try:
            await self.consumer.start()
            self._running = True
            logger.info(
                f"Kafka Consumer started on brokers: {self.consumer._client.hosts}")
            logger.info(
                f"Subscribed to topics: {self.consumer.subscription()}")

            async for msg in self.consumer:
                if not self._running:
                    break
                await self.logs_service.process_event(msg.value)

        except Exception as e:
            logger.error(f"Error in Kafka Consumer loop: {e}")
        finally:
            await self.stop()

    async def stop(self):
        if self._running:
            self._running = False
            await self.consumer.stop()
            logger.info("Kafka Consumer connection closed")
