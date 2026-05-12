package usecase

import (
	"KolinFinance/internal/domain"
	"KolinFinance/internal/mocks"
	"KolinFinance/internal/pkg/logger"
	"context"
	"errors"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
	// gomock — библиотека для создания и управления моками
	// testify/assert — удобные проверки вместо if err != nil { t.Fatalf(...) }
)

// конструктор лля тестов
// логгер с уровнем ерор чтобы вывод был максимально коротким
func newTestUsecase(repo *mocks.MockTransactionRepository) *transactionUsecase {
	return &transactionUsecase{
		repo: repo,
		log:  logger.New(logger.LevelError),
	}
}

// ── Add ──────────────────────────────────────────────────────────────────────

func TestAdd_Success(t *testing.T) {
	// Arrange

	// gomock.NewController — создаёт контроллер который следит за моками.
	// t.Cleanup гарантирует что ctrl.Finish() вызовется в конце теста.
	// Finish проверяет что все ожидаемые вызовы действительно произошли.
	ctrl := gomock.NewController(t)

	// mocks.NewMockTransactionRepository — сгенерированный мок.
	// Реализует ports.TransactionRepository автоматически.
	repo := mocks.NewMockTransactionRepository(ctrl)

	want := domain.Transaction{
		ID:       1,
		Type:     domain.Income,
		Amount:   5000,
		Category: domain.Salary,
	}

	// EXPECT() — говорим моку что мы ожидаем от него.
	// Save должен быть вызван ровно один раз (Times(1))
	// с любым контекстом (gomock.Any()) и любой транзакцией (gomock.Any()).
	// Return — что вернуть когда вызов произойдёт.
	repo.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Times(1).
		Return(want, nil)

	uc := newTestUsecase(repo)
	tx := domain.Transaction{
		Type:     domain.Income,
		Amount:   5000,
		Category: domain.Salary,
	}

	// Act
	saved, err := uc.Add(context.Background(), tx)

	// Assert

	// assert.NoError — проверяет что ошибки нет.
	// Эквивалент if err != nil { t.Fatalf(...) } но короче и читаемее.
	assert.NoError(t, err)

	// assert.Equal — проверяет равенство.
	// Первый аргумент — ожидаемое, второй — полученное.
	assert.Equal(t, want.ID, saved.ID)

	// assert.False — проверяет что условие ложно.
	assert.False(t, saved.Date.IsZero(), "дата должна быть заполнена автоматически")
}

func TestAdd_InvalidAmount(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTransactionRepository(ctrl)

	// Save не должен вызываться — валидация должна упасть раньше.
	// Times(0) — явно говорим что вызовов не ожидаем.
	// Если Save всё же вызовется — тест упадёт.
	repo.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Times(0)

	uc := newTestUsecase(repo)
	tx := domain.Transaction{
		Type:     domain.Income,
		Amount:   0,
		Category: domain.Salary,
	}

	// Act
	_, err := uc.Add(context.Background(), tx)

	// Assert
	// assert.Error — проверяет что ошибка есть
	assert.Error(t, err)
	// assert.ErrorIs — аналог errors.Is, проверяет цепочку оборачивания
	assert.ErrorIs(t, err, domain.ErrIvalidAmount)
}

func TestAdd_InvalidType(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTransactionRepository(ctrl)
	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)

	uc := newTestUsecase(repo)
	tx := domain.Transaction{
		Type:     "неизвестный",
		Amount:   1000,
		Category: domain.Salary,
	}

	// Act
	_, err := uc.Add(context.Background(), tx)

	// Assert
	assert.ErrorIs(t, err, domain.ErrInvalidType)
}

func TestAdd_InvalidCategory(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTransactionRepository(ctrl)
	repo.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)

	uc := newTestUsecase(repo)
	tx := domain.Transaction{
		Type:     domain.Income,
		Amount:   1000,
		Category: "несуществующая",
	}

	// Act
	_, err := uc.Add(context.Background(), tx)

	// Assert
	assert.ErrorIs(t, err, domain.ErrInvalidCategory)
}

func TestAdd_RepoError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTransactionRepository(ctrl)

	repoErr := errors.New("база упала")
	repo.EXPECT().
		Save(gomock.Any(), gomock.Any()).
		Times(1).
		Return(domain.Transaction{}, repoErr)

	uc := newTestUsecase(repo)
	tx := domain.Transaction{
		Type:     domain.Expense,
		Amount:   100,
		Category: domain.Food,
	}

	// Act
	_, err := uc.Add(context.Background(), tx)

	// Assert
	assert.ErrorIs(t, err, repoErr)
}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestGetByID_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTransactionRepository(ctrl)

	want := domain.Transaction{
		ID:       42,
		Type:     domain.Expense,
		Amount:   300,
		Category: domain.Food,
	}

	// gomock.Any() для контекста, конкретное значение для id —
	// проверяем что usecase передаёт правильный id в репозиторий
	repo.EXPECT().
		FindByID(gomock.Any(), int64(42)).
		Times(1).
		Return(want, nil)

	uc := newTestUsecase(repo)

	// Act
	got, err := uc.GetByID(context.Background(), 42)

	// Assert
	assert.NoError(t, err)
	assert.Equal(t, want.ID, got.ID)
	assert.Equal(t, want.Amount, got.Amount)
}

func TestGetByID_NotFound(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTransactionRepository(ctrl)

	repo.EXPECT().
		FindByID(gomock.Any(), int64(99)).
		Times(1).
		Return(domain.Transaction{}, domain.ErrNotFound)

	uc := newTestUsecase(repo)

	// Act
	_, err := uc.GetByID(context.Background(), 99)

	// Assert
	assert.ErrorIs(t, err, domain.ErrNotFound)
}

// ── GetAll ────────────────────────────────────────────────────────────────────

func TestGetAll_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTransactionRepository(ctrl)

	want := []domain.Transaction{
		{ID: 1, Type: domain.Income, Amount: 5000, Category: domain.Salary},
		{ID: 2, Type: domain.Expense, Amount: 200, Category: domain.Food},
	}
	repo.EXPECT().
		FindAll(gomock.Any()).
		Times(1).
		Return(want, nil)

	uc := newTestUsecase(repo)

	// Act
	got, err := uc.GetAll(context.Background())

	// Assert
	assert.NoError(t, err)
	// assert.Len — проверяет длину слайса
	assert.Len(t, got, len(want))
}

func TestGetAll_RepoError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTransactionRepository(ctrl)

	repoErr := errors.New("база упала")
	repo.EXPECT().
		FindAll(gomock.Any()).
		Times(1).
		Return(nil, repoErr)

	uc := newTestUsecase(repo)

	// Act
	_, err := uc.GetAll(context.Background())

	// Assert
	assert.ErrorIs(t, err, repoErr)
}

// ── GetByType ─────────────────────────────────────────────────────────────────

func TestGetByType_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTransactionRepository(ctrl)

	want := []domain.Transaction{
		{ID: 1, Type: domain.Expense, Amount: 200, Category: domain.Food},
	}

	// проверяем что в репо передаётся конкретный тип
	repo.EXPECT().
		FindByType(gomock.Any(), domain.Expense).
		Times(1).
		Return(want, nil)

	uc := newTestUsecase(repo)

	// Act
	got, err := uc.GetByType(context.Background(), domain.Expense)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, got, 1)
	assert.Equal(t, domain.Expense, got[0].Type)
}

// ── GetByPeriod ───────────────────────────────────────────────────────────────

func TestGetByPeriod_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTransactionRepository(ctrl)

	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)
	want := []domain.Transaction{
		{ID: 1, Type: domain.Income, Amount: 5000, Category: domain.Salary, Date: from},
	}

	repo.EXPECT().
		FindByPeriod(gomock.Any(), from, to).
		Times(1).
		Return(want, nil)

	uc := newTestUsecase(repo)

	// Act
	got, err := uc.GetByPeriod(context.Background(), from, to)

	// Assert
	assert.NoError(t, err)
	assert.Len(t, got, len(want))
}

func TestGetByPeriod_InvalidRange(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTransactionRepository(ctrl)

	// FindByPeriod не должен вызываться — проверка периода раньше
	repo.EXPECT().
		FindByPeriod(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(0)

	uc := newTestUsecase(repo)
	from := time.Now()
	to := from.Add(-24 * time.Hour)

	// Act
	_, err := uc.GetByPeriod(context.Background(), from, to)

	// Assert
	assert.Error(t, err)
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestDelete_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTransactionRepository(ctrl)

	repo.EXPECT().
		Delete(gomock.Any(), int64(1)).
		Times(1).
		Return(nil)

	uc := newTestUsecase(repo)

	// Act
	err := uc.Delete(context.Background(), 1)

	// Assert
	assert.NoError(t, err)
}

func TestDelete_NotFound(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	repo := mocks.NewMockTransactionRepository(ctrl)

	repo.EXPECT().
		Delete(gomock.Any(), int64(99)).
		Times(1).
		Return(domain.ErrNotFound)

	uc := newTestUsecase(repo)

	// Act
	err := uc.Delete(context.Background(), 99)

	// Assert
	assert.ErrorIs(t, err, domain.ErrNotFound)
}
