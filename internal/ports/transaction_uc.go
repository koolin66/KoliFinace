package ports

import (
	"KolinFinance/internal/domain"
	"context"
	"time"
)

// контракт бизнес-логики для транзакций
// (что умееет делать usecase с моими транзами)
type TransactionUsecase interface {
	// валидирует и сохраняет новую транзакцию
	Add(ctx context.Context, tx domain.Transaction) (domain.Transaction, error)

	// возвращает транзакцию по ID
	GetByID(ctx context.Context, id int64) (domain.Transaction, error)

	// возвращает все транзакции
	GetAll(ctx context.Context) ([]domain.Transaction, error)

	// возвращает только доходы или только расходы
	GetByType(ctx context.Context, t domain.Type) ([]domain.Transaction, error)

	// возвращает транзы за определенный период
	GetByPeriod(ctx context.Context, from, to time.Time) ([]domain.Transaction, error)

	// удаляет транзакцию по ID
	Delete(ctx context.Context, id int64) error
}
