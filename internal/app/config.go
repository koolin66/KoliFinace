package app

import (
	"KolinFinance/internal/api/delivery"
	"KolinFinance/internal/infrastructure/pg"
	"KolinFinance/internal/pkg/logger"
	"log"

	"github.com/joho/godotenv"
	"github.com/kelseyhightower/envconfig"
)

type Config struct {
	ServerConfig delivery.ServerConfig `env:"SERVER"`
	DB           pg.Config             `env:"DB"`
	LoggerLevel  logger.Level          `env:"LOG"`
}

// загружает конфиг: подтягивает .env (godotenv), затем заполняет структуру из окружения (envconfig)
func LoadConfig() (Config, error) {
	if err := godotenv.Load("/Users/nikolaytrusov/KolinFinance/cmd/finance/.env"); err != nil {
		log.Printf("config: .env не найден, используем окружение: %v", err)
	}

	var cfg Config
	if err := envconfig.Process("APP", &cfg); err != nil {
		return Config{}, err
	}
	return cfg, nil

}
