package config

import (
	"log"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	SrvAddress string `yaml:"srv_address" env:"SERVER_ADDRESS" end-default:"localhost:8080"`
	BaseURL    string `yaml:"base_url" env:"BASE_URL" end-default:"localhost:8080"`
	DBAddress  string `yaml:"db_address" env:"DB_ADDRESS" end-default:"localhost:8080"`
	RedisURL   string `yaml:"redis_url" env:"REDIS_URL" end-default:"localhost:8080"`
}

func MustLoad(configPath string) Config {
	var cfg Config
	if err := cleanenv.ReadConfig(configPath, &cfg); err != nil {
		log.Fatalf("cannot read config %q: %s", configPath, err)
	}
	return cfg
}
