from typing import List
from pydantic_settings import BaseSettings, SettingsConfigDict
from pydantic import Field
# , computed_field


class Settings(BaseSettings):
    POSTGRES_HOST: str = Field(default=..., alias="LOGS_POSTGRES_HOST")
    POSTGRES_USER: str = Field(default=..., alias="LOGS_POSTGRES_USER")
    POSTGRES_PASSWORD: str = Field(default=..., alias="LOGS_POSTGRES_PASSWORD")
    POSTGRES_DB: str = Field(default=..., alias="LOGS_POSTGRES_DB")
    POSTGRES_PORT: int = Field(default=..., alias="LOGS_POSTGRES_PORT")

    BROKERS: str = Field(default=..., alias="KAFKA_BOOTSTRAP_SERVERS")
    AUTH_EVENTS_TOPIC: str = Field(
        default=..., alias="KAFKA_AUTH_EVENTS_TOPIC")
    COMPANY_EVENTS_TOPIC: str = Field(
        default=..., alias="KAFKA_COMPANY_EVENTS_TOPIC")
    FILE_EVENTS_TOPIC: str = Field(
        default=..., alias="KAFKA_FILE_EVENTS_TOPIC")

    # @computed_field
    @property
    def TOPIC(self) -> List[str]:
        return [self.COMPANY_EVENTS_TOPIC, self.AUTH_EVENTS_TOPIC, self.FILE_EVENTS_TOPIC]

    model_config = SettingsConfigDict(
        env_file_encoding="utf-8",
        extra="ignore"
    )


settings = Settings()
