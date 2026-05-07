package app

import (
	"KolinFinance/internal/api/delivery"
	"KolinFinance/internal/infrastructure/pg"
	"KolinFinance/internal/pkg/logger"
	"KolinFinance/internal/usecase"
	"context"
	"fmt"
	"net/http"
	"time"

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

	//Logger--------------------------------------------------

	log := logger.New(cfg.LogLevel)
	//SQL---------------------------------------------------
	pool, err := pgxpool.New(context.Background(), cfg.DatabaseURL)
	if err != nil {
		return nil, fmt.Errorf("ошибка подключения к постгрес", "err", err)
	}

	if err := pool.Ping(context.Background()); err != nil {
		return nil, fmt.Errorf("ошибка проверки подключения к посгрес", "err", err)
	}

	log.Info("подкючение с посгре установлено")

	//Migrate--------------------------------------------------

	if err := pg.Migrate(context.Background(), pool); err != nil {
		return nil, fmt.Errorf("ошибка создания таблицы в бд", "err", err)
	}
	log.Info("таблица создана")

	//Repository---------------------------------------------------

	repo := pg.NewPostgresRepo(pool)

	//Usecases---------------------------------------------------

	txUC := usecase.NewTransactionUsecase(repo, log)
	reportUC := usecase.NewReportUsecase(repo, log)

	//Handler---------------------------------------------------

	handler := delivery.NewHandler(txUC, reportUC, log)

	//http server---------------------------------------------------

	server := &http.Server{
		Addr:         ":" + cfg.HTTPPort,
		Handler:      handler.Router(),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
		//idle время keep alive пока ждет ноый запрос
		IdleTimeout: 60 * time.Second,
	}

	return &App{server: server, log: log}, nil
	//---------------------------------------------------

}

func (a *App) Run() error {
	a.log.Info("сервер запущен", "addr", a.server.Addr)
	return a.server.ListenAndServe()
}
