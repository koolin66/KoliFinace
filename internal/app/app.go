package app

import (
	"KolinFinance/internal/api/delivery"
	"KolinFinance/internal/api/grpc"
	"KolinFinance/internal/infrastructure/pg"
	"KolinFinance/internal/pkg/logger"
	"KolinFinance/internal/usecase"
	transaction "KolinFinance/proto"
	"context"
	"fmt"
	"net"
	"net/http"
	"time"

	"google.golang.org/grpc/reflection"

	"github.com/jackc/pgx/v5/pgxpool"
	g "google.golang.org/grpc"
)

type App struct {
	server *http.Server
	log    *logger.Logger
}

func New(cfg Config) (*App, error) {

	//Logger--------------------------------------------------

	log := logger.New(cfg.LoggerLevel)
	//SQL---------------------------------------------------

	// собираю строку подключения к бд
	dsn := fmt.Sprintf(
		"postgres://%s:%s@%s:%s/%s",
		cfg.DB.User,
		cfg.DB.Password,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.DBName,
	)
	//это подключение к бд, у Калькулятора это
	pool, err := pgxpool.New(context.Background(), dsn)
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

	//------gRPC server -------------------------------------------------
	grpcHandler := grpc.NewHandler(txUC, reportUC)
	grpcServer := g.NewServer()
	transaction.RegisterTransactionServiceServer(grpcServer, grpcHandler)
	reflection.Register(grpcServer)

	lis, err := net.Listen("tcp", ":9090")
	if err != nil {
		return nil, fmt.Errorf("ошибка запуска grpc listener: %w", err)
	}

	go func() {
		if err := grpcServer.Serve(lis); err != nil {
			log.Error("grpc server error", "err", err)
		}
	}()

	log.Info("grpc сервер запущен", "port", "9090")
	//http server---------------------------------------------------

	server := &http.Server{
		Addr:         ":" + cfg.ServerConfig.Port,
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
