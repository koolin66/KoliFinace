package usecase

import (
	"KolinFinance/internal/domain"
	"context"
	"time"
)

// пишу ручные моки
type mockRepo struct {

	//моки на методы репозитория
	//Fn значит в значении function
	saveFn         func(ctx context.Context, tx domain.Transaction) (domain.Transaction, error)
	findByIDFn     func(ctx context.Context, id int64) (domain.Transaction, error)
	findAllFn      func(ctx context.Context) ([]domain.Transaction, error)
	findByTypeFn   func(ctx context.Context, t domain.Type) ([]domain.Transaction, error)
	findByPeriodFn func(ctx context.Context, from, to time.Time) ([]domain.Transaction, error)
	deleteFn       func(ctx context.Context, id int64) error
}

func (m *mockRepo) Save(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	return m.saveFn(ctx, tx)
}

func (m *mockRepo) FindByID(ctx context.Context, id int64) (domain.Transaction, error) {
	return m.findByIDFn(ctx, id)
}

func (m *mockRepo) FindAll(ctx context.Context) ([]domain.Transaction, error) {
	return m.findAllFn(ctx)
}

func (m *mockRepo) FindByType(ctx context.Context, t domain.Type) ([]domain.Transaction, error) {
	return m.findByTypeFn(ctx, t)
}

func (m *mockRepo) FindByPeriod(ctx context.Context, from, to time.Time) ([]domain.Transaction, error) {
	return m.findByPeriodFn(ctx, from, to)
}

func (m *mockRepo) Delete(ctx context.Context, id int64) error {
	return m.deleteFn(ctx, id)
}
