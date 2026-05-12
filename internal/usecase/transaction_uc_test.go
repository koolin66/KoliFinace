package usecase

import (
	"KolinFinance/internal/domain"
	"KolinFinance/internal/mocks"
	"KolinFinance/internal/pkg/logger"
	"context"
	"testing"

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

func TestAdd_Success(t *testing.T) {
	//arrange

	// gomock.NewController — создаёт контроллер который следит за моками.
	// t.Cleanup гарантирует что ctrl.Finish() вызовется в конце теста.
	// Finish проверяет что все ожидаемые вызовы действительно произошли.
	ctrl := gomock.NewController(t)

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
	repo.EXPECT().Save(gomock.Any(), gomock.Any()).
		Times(1).
		Return(want, nil)

	uc := newTestUsecase(repo)
	tx := domain.Transaction{
		Type:     domain.Income,
		Amount:   5000,
		Category: domain.Salary,
	}
	//act
	saved, err := uc.Add(context.Background(), tx)

	//assert

	//короткая проверка ошибки, аналог if err != nil { t.Fatalf(...) }
	assert.NoError(t, err)

	// сравнение
	assert.Equal(t, want.ID, saved.ID)

	//проверяяет что условие ложно
	assert.False(t, saved.Date.IsZero(), "дата должна быть заполнена автоматически")
}
