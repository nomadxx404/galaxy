package config

import (
	"fmt"
	"log"
	"net/url"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Postgres PostgresConfig
	Redis    RedisConfig
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

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		instance = &Config{}
		_ = cleanenv.ReadConfig(".env", instance)

		if err := cleanenv.ReadEnv(instance); err != nil {
			log.Printf("failed to read env: %v", err)
		}
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
