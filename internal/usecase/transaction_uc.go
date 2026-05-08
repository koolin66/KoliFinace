package usecase

import (
	"KolinFinance/internal/domain"
	"KolinFinance/internal/pkg/logger"
	"KolinFinance/internal/ports"
	"context"
	"fmt"
	"time"
)

// снаружи никто не знает о этой структуре у которой есть все методы что ниже, потому что
// она закидывается на ручку интерфейса, поэтому сам юзкейс мы может в будущем менять
type transactionUsecase struct {
	repo ports.TransactionRepository
	log  *logger.Logger
}

// бля, тут короче присоединяем наш репозиторий через интерфейс, и возвращаем эту сборку в виде интерфейса для юзкейса
// а юзкейс будет присоединять в другом месте, скорее все в апп
func NewTransactionUsecase(repo ports.TransactionRepository, log *logger.Logger) ports.TransactionUsecase {
	return &transactionUsecase{repo: repo, log: log}
}

// это логика методов для юзкейса, внутри спрятаны ручки интерфейса для репозитория
func (uc *transactionUsecase) Add(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
	if tx.Date.IsZero() {
		tx.Date = time.Now()
	}

	if err := tx.Validate(); err != nil {
		uc.log.Warn("транзакция не прошла валидацию", err)
		return domain.Transaction{}, fmt.Errorf("ошибка валидации транзы", err)
	}

	saved, err := uc.repo.Save(ctx, tx)
	if err != nil {
		uc.log.Error("ошибка сохранения", err)
		return domain.Transaction{}, fmt.Errorf("ошибка сохранения транзакции", err)
	}
	uc.log.Info("транзкация сохранена", "id", saved.ID)
	return saved, nil
}

func (uc *transactionUsecase) GetByID(ctx context.Context, id int64) (domain.Transaction, error) {
	tx, err := uc.repo.FindByID(ctx, id)
	if err != nil {
		uc.log.Warn("транзакция не найдена", "id", id, "err", err)
		return domain.Transaction{}, fmt.Errorf("транзакция не найдена %w", err)
	}
	return tx, nil
}

func (uc *transactionUsecase) GetAll(ctx context.Context) ([]domain.Transaction, error) {
	txs, err := uc.repo.FindAll(ctx)
	if err != nil {
		uc.log.Error("ошибка получения транзакций", "err", err)
		return nil, err
	}
	return txs, nil
}

func (uc *transactionUsecase) GetByType(ctx context.Context, t domain.Type) ([]domain.Transaction, error) {
	txs, err := uc.repo.FindByType(ctx, t)
	if err != nil {
		uc.log.Error("не получилось получить транзакции по типу", "тип", t, "err", err)
		return nil, fmt.Errorf("ошибка получения по типу: %w", err)
	}
	return txs, nil
}

func (uc *transactionUsecase) GetByPeriod(ctx context.Context, from, to time.Time) ([]domain.Transaction, error) {
	if from.After(to) {
		return nil, fmt.Errorf("невалидный период: начало %v позже конца %v", from, to)
	}

	txs, err := uc.repo.FindByPeriod(ctx, from, to)
	if err != nil {
		uc.log.Error("транзакции за данный период не найдены", "от", from, "до", to)
		return nil, fmt.Errorf("транзакции за данный период не найдены: %w", err)
	}
	return txs, nil
}

func (uc *transactionUsecase) Delete(ctx context.Context, id int64) error {
	if err := uc.repo.Delete(ctx, id); err != nil {
		uc.log.Warn("ошибка удаления транзы", "id", id, "err", err)
		return fmt.Errorf("ошибка удаления тразакции %w", err)
	}
	uc.log.Info("транзакция успешно удалена", "id", id)
	return nil
}
