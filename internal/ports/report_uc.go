package ports

import (
	"KolinFinance/internal/domain"
	"context"
	"time"
)

// сводка по одной категории
type Summary struct {
	Category domain.Category
	Total    float64
	Count    int
}

// полный фин отчет за определенный период
type Report struct {
	From      time.Time
	To        time.Time
	Income    float64
	Expense   float64
	Balance   float64
	Summaries []Summary
}

//тут уже суета, также для usecase, но только по функционалу
//для отчетов о финансах (агрегированные данные)

// что должен уметь делать с данными usecase?
// вот что:
type ReportUseCase interface {
	// текущий баланс: доходы - расходы
	Balance(ctx context.Context) (float64, error)
	// отчет за указаный период
	ReportByPeriod(ctx context.Context, from, to time.Time) (Report, error)
}
