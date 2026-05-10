package config

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Postgres PostgresConfig
	Kafka    KafkaConfig
	Mimio    MinioConfig
	BaseUrl  BaseUrlConfig
}

type PostgresConfig struct {
	POSTGRES_HOST     string `env:"POSTGRES_HOST"`
	POSTGRES_USER     string `env:"POSTGRES_USER"`
	POSTGRES_PASSWORD string `env:"POSTGRES_PASSWORD"`
	POSTGRES_DB       string `env:"POSTGRES_DB"`
	POSTGRES_PORT     string `env:"POSTGRES_PORT"`
}

type KafkaConfig struct {
	BROKERS              string `env:"KAFKA_BOOTSTRAP_SERVERS"`
	EVENTS_TOPIC_AUTH    string `env:"KAFKA_AUTH_EVENTS_TOPIC"`
	EVENTS_TOPIC_COMPANY string `env:"KAFKA_COMPANY_EVENTS_TOPIC"`
	EVENTS_TOPIC_FILE    string `env:"KAFKA_FILE_EVENTS_TOPIC"`
}

type MinioConfig struct {
	Host         string `env:"MINIO_HOST"`
	Port         string `env:"MINIO_PORT"`
	RootUser     string `env:"MINIO_ROOT_USER"`
	RootPassword string `env:"MINIO_ROOT_PASSWORD"`
	BucketName   string `env:"MINIO_BUCKET_NAME"`
}

type BaseUrlConfig struct {
	FILES_SERVICE_URL string `env:"FILES_SERVICE_URL"`
}

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		_ = cleanenv.ReadConfig(".env", instance)
		cleanenv.ReadEnv(instance)
	})
	return instance
}

func (pc PostgresConfig) GetDSN() string {
	escapedPassword := url.PathEscape(pc.POSTGRES_PASSWORD)

	return fmt.Sprintf("postgres://%s:%s@%s:%s/%s?sslmode=disable",
		pc.POSTGRES_USER,
		escapedPassword,
		pc.POSTGRES_HOST,
		pc.POSTGRES_PORT,
		pc.POSTGRES_DB,
	)
}
