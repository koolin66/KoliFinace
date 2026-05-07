package main

import (
	"KolinFinance/internal/app"
	"KolinFinance/internal/pkg/logger"
	"errors"
	"log"
	"net/http"
	"os"
)

func main() {
	cfg := app.Config{ //string conn к postgresql ^
		// протокол://пользователь:пароль@хост:порт/база_данных
		DatabaseURL: getEnv("DATABASE_URL", "postgres://postgres:postgres@localhost:5432/finance"),
		HTTPPort:    getEnv("HTTP_PORT", "8080"),
		LogLevel:    logger.Level(getEnv("LOG_LEVEL", "info")),
	}

	application, err := app.New(cfg)
	if err != nil {
		log.Fatalf("ошибка создания программы %v", err)
	}

	if err := application.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatalf("ошибка сервера %v", err)
	}
}

// getEnv с fallback позволяет запускать без настройки
// переменных окружения — дефолты работают из коробки для
// локальной разработки.

// эта функция читает переменную окружения, если ее нет,
// выставляет дефолтные значения
func getEnv(key, fallback string) string {
	if value, ok := os.LookupEnv(key); ok {
		return value
	}
	return fallback
}
