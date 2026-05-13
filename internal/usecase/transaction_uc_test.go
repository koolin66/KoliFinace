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

func TestAdd(t *testing.T) {

	tests := []struct {
		name      string
		tx        domain.Transaction
		mockSetup func(repo *mocks.MockTransactionRepository)
		wantErr   error
		wantID    int64
	}{
		{
			name: "успешное добавление",
			tx: domain.Transaction{
				Date:     time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC),
				Type:     domain.Income,
				Amount:   5000,
				Category: domain.Salary,
			},
			mockSetup: func(repo *mocks.MockTransactionRepository) {
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Times(1).Return(
					domain.Transaction{ID: 1, Date: time.Date(2024, 1, 15, 0, 0, 0, 0, time.UTC), Type: domain.Income, Amount: 5000, Category: domain.Salary}, nil)

			},
			wantErr: nil,
			wantID:  1,
		},
		{
			// валидация должна упасть до обращения к репо
			name: "невалидная сумма",
			tx: domain.Transaction{
				Type:     domain.Income,
				Amount:   0,
				Category: domain.Salary,
			},
			mockSetup: func(repo *mocks.MockTransactionRepository) {
				// Save не должен вызываться — Times(0) это гарантирует
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
			},
			wantErr: domain.ErrIvalidAmount,
		},
		{
			name: "невалидный тип",
			tx: domain.Transaction{
				Type:     "неизвестный",
				Amount:   1000,
				Category: "food",
			},
			mockSetup: func(repo *mocks.MockTransactionRepository) {
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
			},
			wantErr: domain.ErrInvalidType,
		},
		{
			name: "невалидная категория",
			tx: domain.Transaction{
				Type:     "income",
				Amount:   5000,
				Category: "abc",
			},
			mockSetup: func(repo *mocks.MockTransactionRepository) {
				repo.EXPECT().Save(gomock.Any(), gomock.Any()).Times(0)
			},
			wantErr: domain.ErrInvalidCategory,
		},
		{
			name: "ошибка репозитория",
			tx: domain.Transaction{
				Type:     domain.Expense,
				Amount:   100,
				Category: domain.Food,
			},
			mockSetup: func(repo *mocks.MockTransactionRepository) {
				repo.EXPECT().
					Save(gomock.Any(), gomock.Any()).
					Times(1).
					Return(domain.Transaction{}, errors.New("база упала"))
			},
			wantErr: errors.New("база упала"),
		},
	}

	// t.Run запускает подтест с именем tt.name.
	// В выводе go test -v будет видно: TestAdd/успешное_добавление,
	// TestAdd/невалидная_сумма и так далее.
	// Каждый подтест независим — падение одного не останавливает остальные.
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			//arrange
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockTransactionRepository(ctrl)

			//настройка мока для каждого сценария
			tt.mockSetup(repo)

			uc := newTestUsecase(repo)

			//act

			saved, err := uc.Add(context.Background(), tt.tx)

			//accert
			if tt.wantErr != nil {
				// ожидаем ошибку — проверяем что она есть и правильная
				//assert.Error(t, err) проверяет, что ошибка существует (т.е. err != nil).
				assert.Error(t, err)
				if errors.Is(tt.wantErr, domain.ErrIvalidAmount) ||
					errors.Is(tt.wantErr, domain.ErrInvalidType) ||
					errors.Is(tt.wantErr, domain.ErrInvalidCategory) {
					assert.ErrorIs(t, err, tt.wantErr)
				}
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantID, saved.ID)
				assert.False(t, saved.Date.IsZero(), "дата должна быть заполнена автоматически")
			}

		})
	}

}

// ── GetByID ───────────────────────────────────────────────────────────────────

func TestGetByID(t *testing.T) {
	// для GetByID нам нужно передавать id в мок —
	// добавляем поле inputID в таблицу
	tests := []struct {
		name      string
		inputID   int64
		mockSetup func(repo *mocks.MockTransactionRepository)
		wantErr   error
		wantTx    domain.Transaction
	}{
		{
			name:    "успешное получение",
			inputID: 42,
			mockSetup: func(repo *mocks.MockTransactionRepository) {
				repo.EXPECT().
					FindByID(gomock.Any(), int64(42)).
					Times(1).
					Return(domain.Transaction{ID: 42, Type: domain.Expense, Amount: 300, Category: domain.Food}, nil)
			},
			wantErr: nil,
			wantTx:  domain.Transaction{ID: 42, Amount: 300},
		},
		{
			name:    "не найдено",
			inputID: 99,
			mockSetup: func(repo *mocks.MockTransactionRepository) {
				repo.EXPECT().
					FindByID(gomock.Any(), int64(99)).
					Times(1).
					Return(domain.Transaction{}, domain.ErrNotFound)
			},
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockTransactionRepository(ctrl)
			tt.mockSetup(repo)
			uc := newTestUsecase(repo)

			// Act
			got, err := uc.GetByID(context.Background(), tt.inputID)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.wantTx.ID, got.ID)
				assert.Equal(t, tt.wantTx.Amount, got.Amount)
			}
		})
	}
}

// ── GetAll ────────────────────────────────────────────────────────────────────

func TestGetAll(t *testing.T) {
	// wantLen — сколько транзакций ожидаем получить
	tests := []struct {
		name      string
		mockSetup func(repo *mocks.MockTransactionRepository)
		wantErr   error
		wantLen   int
	}{
		{
			name: "успешное получение всех",
			mockSetup: func(repo *mocks.MockTransactionRepository) {
				repo.EXPECT().
					FindAll(gomock.Any()).
					Times(1).
					Return([]domain.Transaction{
						{ID: 1, Type: domain.Income, Amount: 5000, Category: domain.Salary},
						{ID: 2, Type: domain.Expense, Amount: 200, Category: domain.Food},
					}, nil)
			},
			wantLen: 2,
		},
		{
			name: "пустой список",
			mockSetup: func(repo *mocks.MockTransactionRepository) {
				repo.EXPECT().
					FindAll(gomock.Any()).
					Times(1).
					Return([]domain.Transaction{}, nil)
			},
			wantLen: 0,
		},
		{
			name: "ошибка репозитория",
			mockSetup: func(repo *mocks.MockTransactionRepository) {
				repo.EXPECT().
					FindAll(gomock.Any()).
					Times(1).
					Return(nil, errors.New("база упала"))
			},
			wantErr: errors.New("база упала"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockTransactionRepository(ctrl)
			tt.mockSetup(repo)
			uc := newTestUsecase(repo)

			// Act
			got, err := uc.GetAll(context.Background())

			// Assert
			if tt.wantErr != nil {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}

// ── GetByType ─────────────────────────────────────────────────────────────────

func TestGetByType(t *testing.T) {
	tests := []struct {
		name      string
		inputType domain.Type
		mockSetup func(repo *mocks.MockTransactionRepository)
		wantErr   error
		wantLen   int
	}{
		{
			name:      "только расходы",
			inputType: domain.Expense,
			mockSetup: func(repo *mocks.MockTransactionRepository) {
				repo.EXPECT().
					FindByType(gomock.Any(), domain.Expense).
					Times(1).
					Return([]domain.Transaction{
						{ID: 1, Type: domain.Expense, Amount: 200, Category: domain.Food},
						{ID: 2, Type: domain.Expense, Amount: 150, Category: domain.Transport},
					}, nil)
			},
			wantLen: 2,
		},
		{
			name:      "только доходы",
			inputType: domain.Income,
			mockSetup: func(repo *mocks.MockTransactionRepository) {
				repo.EXPECT().
					FindByType(gomock.Any(), domain.Income).
					Times(1).
					Return([]domain.Transaction{
						{ID: 3, Type: domain.Income, Amount: 5000, Category: domain.Salary},
					}, nil)
			},
			wantLen: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockTransactionRepository(ctrl)
			tt.mockSetup(repo)
			uc := newTestUsecase(repo)

			// Act
			got, err := uc.GetByType(context.Background(), tt.inputType)

			// Assert
			assert.NoError(t, err)
			assert.Len(t, got, tt.wantLen)
			// проверяем что все транзакции нужного типа
			for _, tx := range got {
				assert.Equal(t, tt.inputType, tx.Type)
			}
		})
	}
}

// ── GetByPeriod ───────────────────────────────────────────────────────────────

func TestGetByPeriod(t *testing.T) {
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	tests := []struct {
		name      string
		from      time.Time
		to        time.Time
		mockSetup func(repo *mocks.MockTransactionRepository)
		wantErr   bool
		wantLen   int
	}{
		{
			name: "успешное получение за период",
			from: from,
			to:   to,
			mockSetup: func(repo *mocks.MockTransactionRepository) {
				repo.EXPECT().
					FindByPeriod(gomock.Any(), from, to).
					Times(1).
					Return([]domain.Transaction{
						{ID: 1, Type: domain.Income, Amount: 5000, Category: domain.Salary, Date: from},
					}, nil)
			},
			wantLen: 1,
		},
		{
			// бизнес-правило: from не может быть позже to
			// репо не должен вызываться
			name: "невалидный период",
			from: to, // переворачиваем — from позже to
			to:   from,
			mockSetup: func(repo *mocks.MockTransactionRepository) {
				repo.EXPECT().
					FindByPeriod(gomock.Any(), gomock.Any(), gomock.Any()).
					Times(0)
			},
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockTransactionRepository(ctrl)
			tt.mockSetup(repo)
			uc := newTestUsecase(repo)

			// Act
			got, err := uc.GetByPeriod(context.Background(), tt.from, tt.to)

			// Assert
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Len(t, got, tt.wantLen)
			}
		})
	}
}

// ── Delete ────────────────────────────────────────────────────────────────────

func TestDelete(t *testing.T) {
	tests := []struct {
		name      string
		inputID   int64
		mockSetup func(repo *mocks.MockTransactionRepository)
		wantErr   error
	}{
		{
			name:    "успешное удаление",
			inputID: 1,
			mockSetup: func(repo *mocks.MockTransactionRepository) {
				repo.EXPECT().
					Delete(gomock.Any(), int64(1)).
					Times(1).
					Return(nil)
			},
			wantErr: nil,
		},
		{
			name:    "транзакция не найдена",
			inputID: 99,
			mockSetup: func(repo *mocks.MockTransactionRepository) {
				repo.EXPECT().
					Delete(gomock.Any(), int64(99)).
					Times(1).
					Return(domain.ErrNotFound)
			},
			wantErr: domain.ErrNotFound,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Arrange
			ctrl := gomock.NewController(t)
			repo := mocks.NewMockTransactionRepository(ctrl)
			tt.mockSetup(repo)
			uc := newTestUsecase(repo)

			// Act
			err := uc.Delete(context.Background(), tt.inputID)

			// Assert
			if tt.wantErr != nil {
				assert.ErrorIs(t, err, tt.wantErr)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}
