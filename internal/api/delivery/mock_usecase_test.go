package delivery

import (
	"KolinFinance/internal/domain"
	"KolinFinance/internal/ports"
	"context"
	"time"
)

// пишем моки для юзкейса
type mockTransactionUC struct {
	addFn         func(ctx context.Context, tx domain.Transaction) (domain.Transaction, error)
	getByIDFn     func(ctx context.Context, id int64) (domain.Transaction, error)
	getAllFn      func(ctx context.Context) ([]domain.Transaction, error)
	getByTypeFn   func(ctx context.Context, t domain.Type) ([]domain.Transaction, error)
	getByPeriodFn func(ctx context.Context, from, to time.Time) ([]domain.Transaction, error)
	deleteFn      func(ctx context.Context, id int64) error
}

// пишем методы для интерфейса, чтобы в структуре хендлер могли прокинуть
// свои моки
func (m *mockTransactionUC) Add(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	return m.addFn(ctx, tx)
}
func (m *mockTransactionUC) GetByID(ctx context.Context, id int64) (domain.Transaction, error) {
	return m.getByIDFn(ctx, id)
}
func (m *mockTransactionUC) GetAll(ctx context.Context) ([]domain.Transaction, error) {
	return m.getAllFn(ctx)
}
func (m *mockTransactionUC) GetByType(ctx context.Context, t domain.Type) ([]domain.Transaction, error) {
	return m.getByTypeFn(ctx, t)
}
func (m *mockTransactionUC) GetByPeriod(ctx context.Context, from, to time.Time) ([]domain.Transaction, error) {
	return m.getByPeriodFn(ctx, from, to)
}
func (m *mockTransactionUC) Delete(ctx context.Context, id int64) error {
	return m.deleteFn(ctx, id)
}

// мок для отчетов
type mockReportUC struct {
	balanceFn        func(ctx context.Context) (float64, error)
	reportByPeriodFn func(ctx context.Context, from, to time.Time) (ports.Report, error)
}

func (m *mockReportUC) Balance(ctx context.Context) (float64, error) {
	return m.balanceFn(ctx)
}
func (m *mockReportUC) ReportByPeriod(ctx context.Context, from, to time.Time) (ports.Report, error) {
	return m.reportByPeriodFn(ctx, from, to)
}
