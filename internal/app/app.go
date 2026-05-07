package app

import (
	"/Users/nikolaytrusov/KoliFinance/internal/api/delivery"
	"KolinFinance/internal/infrastructure/pg"
	"KolinFinance/internal/pkg/logger"
	"KolinFinance/internal/usecase"
	"context"
	"fmt"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Config struct {
	DatabaseURL string
	HTTPPort    string
	LogLevel    logger.Level
}

type App struct {
	server *http.Server
	log    *logger.Logger
}

func New(cfg Config) (*App, error) {
	log := logger.New(cfg.LogLevel)
	//---------------------------------------------------
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к постгрес", "err", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ошибка проверки подключения к посгрес", "err", err)
	}

	log.Info("подкючение с посгре установлено")
	//---------------------------------------------------

	if err := pg.Migrate(context.Background(), pool); err != nil {
		return nil, fmt.Errorf("ошибка создания таблицы в бд", "err", err)
	}
	log.Info("таблица создана")
	//---------------------------------------------------
	repo := pg.NewPostgresRepo(pool)
	//---------------------------------------------------

	txUC := usecase.NewTransactionUsecase(repo, log)
	reportUC := usecase.NewReportUsecase(repo, log)

	//---------------------------------------------------

	handler := delivery.NewHandler(txUC, reportUC, log)

	//---------------------------------------------------

}
