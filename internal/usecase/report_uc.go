package usecase

import (
	"KolinFinance/internal/domain"
	"KolinFinance/internal/pkg/logger"
	"KolinFinance/internal/ports"
	"context"
	"fmt"
	"time"
)

type reportUsecase struct {
	repo ports.TransactionRepository
	log  *logger.Logger
}

// все тоже самое что и в транзакциях_юк, это методы для отчетов юзкейса, также имеют в себе
// части интерфейса из репозитория, потому что трогают бд
func NewReportUsecase(repo ports.TransactionRepository, log *logger.Logger) ports.ReportUseCase {
	return &reportUsecase{repo: repo, log: log}
}

func (uc *reportUsecase) Balance(ctx context.Context) (float64, error) {
	txs, err := uc.repo.FindAll(ctx)
	if err != nil {
		uc.log.Error("ошибка получения транзакций для подсчета баланса", "err", err)
		return 0, fmt.Errorf("ошибка получения инфы для подсчета баланса: %w", err)
	}

	var balance float64
	for _, tx := range txs {
		switch tx.Type {
		case domain.Income:
			balance += float64(tx.Amount)
		case domain.Expense:
			balance -= float64(tx.Amount)
		}
	}
	uc.log.Info("баланс посчитан:", "balance:", balance)
	return balance, nil
}

func (uc *reportUsecase) ReportByPeriod(ctx context.Context, from, to time.Time) (ports.Report, error) {
	if from.After(to) {
		return ports.Report{}, fmt.Errorf("invalid period: from %v is after to %v", from, to)
	}

	txs, err := uc.repo.FindByPeriod(ctx, from, to)
	if err != nil {
		uc.log.Error("failed to get transactions for report", "from", from, "to", to, "err", err)
		return ports.Report{}, fmt.Errorf("failed to build report: %w", err)
	}

	report := ports.Report{
		From: from,
		To:   to,
	}

	// считаем суммы по категориям
	categoryTotals := make(map[domain.Category]float64)
	categoryCounts := make(map[domain.Category]int)

	for _, tx := range txs {
		switch tx.Type {
		case domain.Income:
			report.Income += float64(tx.Amount)
		case domain.Expense:
			report.Expense += float64(tx.Amount)
		}
		categoryTotals[tx.Category] += float64(tx.Amount)
		categoryCounts[tx.Category]++
	}

	report.Balance = report.Income - report.Expense

	// собираем срез Summary из map
	for category, total := range categoryTotals {
		report.Summaries = append(report.Summaries, ports.Summary{
			Category: category,
			Total:    total,
			Count:    categoryCounts[category],
		})
	}

	uc.log.Info("report built",
		"from", from,
		"to", to,
		"income", report.Income,
		"expense", report.Expense,
		"balance", report.Balance,
	)

	return report, nil
}
