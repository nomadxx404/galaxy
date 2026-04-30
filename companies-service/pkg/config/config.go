package config

import (
	"fmt"
	"net/url"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Postgres   PostgresConfig
	Redis      RedisConfig
	Kafka      KafkaConfig
	Invitation InvitationConfig
}

type PostgresConfig struct {
	POSTGRES_HOST     string `env:"POSTGRES_HOST"`
	POSTGRES_USER     string `env:"POSTGRES_USER"`
	POSTGRES_PASSWORD string `env:"POSTGRES_PASSWORD"`
	POSTGRES_DB       string `env:"POSTGRES_DB"`
	POSTGRES_PORT     string `env:"POSTGRES_PORT"`
}

type RedisConfig struct {
	REDIS_HOST     string `env:"REDIS_HOST"`
	REDIS_PORT     string `env:"REDIS_PORT"`
	REDIS_PASSWORD string `env:"REDIS_PASSWORD"`
}

type KafkaConfig struct {
	BROKERS           string `env:"KAFKA_BOOTSTRAP_SERVERS"`
	EVENTS_TOPIC      string `env:"KAFKA_COMPANY_EVENTS_TOPIC"`
	EVENTS_TOPIC_AUTH string `env:"KAFKA_AUTH_EVENTS_TOPIC"`
}

type InvitationConfig struct {
	INVITATION_LINK_ACCESS_MINUTES int `env:"INVITATION_LINK_ACCESS_MINUTES"`
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

func (rc RedisConfig) GetAddr() string {
	return fmt.Sprintf("%s:%s", rc.REDIS_HOST, rc.REDIS_PORT)
}
