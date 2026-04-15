package config

import (
	"log"
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	AuthServiceURL    string `env:"AUTH_SERVICE_URL"`
	CompanyServiceURL string `env:"COMPANIES_SERVICE_URL"`
	Port              string `env:"BFF_PORT"`
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
