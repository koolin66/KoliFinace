package ports

import (
	"KolinFinance/internal/domain"
	"context"
	"time"
)

// конракт для работы с хранилищем,
// то есть (методы) всё что можно сделать с доменной сущностью
type TransactionRepository interface {
	//сохраняет новую транзакцию и возвращает ее с заполненым полем id
	//короче, с присвоенным id
	Save(ctx context.Context, tx domain.Transaction) (domain.Transaction, error)

	//возвращает транзакцию из бд по id
	FindByID(ctx context.Context, id int64) (domain.Transaction, error)

	//возвращает все транзакции
	FindAll(ctx context.Context) ([]domain.Transaction, error)

	//возвращает по типу (только доходы или только расходы)
	FindByType(ctx context.Context, t domain.Type) ([]domain.Transaction, error)

	//возвращает транзакции за определенный период
	FindByPeriod(ctx context.Context, from, to time.Time) ([]domain.Transaction, error)

	//удаляет транзу по id
	Delete(ctx context.Context, id int64) error
}
