package grpc

import (
	"KolinFinance/internal/domain"
	"KolinFinance/internal/pkg/logger"
	"KolinFinance/internal/ports"
	transaction "KolinFinance/proto"
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

type Handler struct {
	transaction.UnimplementedTransactionServiceServer
	txUC     ports.TransactionUsecase
	reportUC ports.ReportUseCase
	log      *logger.Logger
}

func NewHandler(txUC ports.TransactionUsecase, reportUC ports.ReportUseCase) *Handler {
	return &Handler{txUC: txUC, reportUC: reportUC, log: &logger.Logger{}}
}

func (h *Handler) Add(ctx context.Context, req *transaction.AddRequest) (*transaction.AddResponse, error) {
	tx := domain.Transaction{
		Type:     domain.Type(req.Type),
		Category: domain.Category(req.Category),
		Amount:   req.Amount,
		Note:     req.Note,
	}

	saved, err := h.txUC.Add(ctx, tx)
	if err != nil {
		if errors.Is(err, domain.ErrIvalidAmount) ||
			errors.Is(err, domain.ErrInvalidType) ||
			errors.Is(err, domain.ErrInvalidCategory) {
			return nil, status.Error(codes.InvalidArgument, "невалидный запрос")
		}
		return nil, status.Error(codes.Internal, "ошибка сервера")
	}

	return &transaction.AddResponse{
		Id:       saved.ID,
		Type:     string(saved.Type),
		Category: string(saved.Category),
		Amount:   saved.Amount,
	}, nil
}

func (h *Handler) GetByID(ctx context.Context, req *transaction.GetByIDRequest) (*transaction.GetByIDResponse, error) {

	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "невалидный ID транзакции")
	}

	tx, err := h.txUC.GetByID(ctx, req.Id)
	if err != nil {
		// 3. Обрабатываем ошибки
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "транзакция не найдена")
		}
		return nil, status.Error(codes.Internal, "ошибка сервера")
	}

	// 4. Формируем успешный ответ
	return &transaction.GetByIDResponse{
		Id:       tx.ID,
		Type:     string(tx.Type),
		Category: string(tx.Category),
		Amount:   tx.Amount,
	}, nil
}

// Delete удаляет транзакцию по ID.
func (h *Handler) Delete(ctx context.Context, req *transaction.DeleteRequest) (*transaction.DeleteResponse, error) {
	// 1. Валидация входных данных
	if req.Id <= 0 {
		return nil, status.Error(codes.InvalidArgument, "невалидный ID транзакции")
	}

	// 2. Вызываем usecase
	err := h.txUC.Delete(ctx, req.Id)
	if err != nil {
		// 3. Обрабатываем ошибки
		if errors.Is(err, domain.ErrNotFound) {
			return nil, status.Error(codes.NotFound, "транзакция не найдена")
		}
		return nil, status.Error(codes.Internal, "ошибка сервера")
	}

	// 4. Успешное удаление (обычно возвращают пустой ответ или ID удалённой записи)
	return &transaction.DeleteResponse{}, nil
}
