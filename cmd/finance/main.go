package main

import (
	"KolinFinance/internal/app"
	"errors"
	"log"
	"net/http"
	"os"
)

func main() {

	cfg, err := app.LoadConfig()
	if err != nil {
		log.Fatalf("ошибка получения конфига %v", err)
		return
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
