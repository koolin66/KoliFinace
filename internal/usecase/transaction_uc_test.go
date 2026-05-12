package usecase

import (
	"KolinFinance/internal/domain"
	"KolinFinance/internal/pkg/logger"
	"context"
	"errors"
	"testing"
	"time"
)

// конструктор лля тестов
// логгер с уровнем ерор чтобы вывод был максимально коротким
func newTestUsecase(repo *mockRepo) *transactionUsecase {
	return &transactionUsecase{
		repo: repo,
		log:  logger.New(logger.LevelError),
	}
}

// метод Add
//

func TestAdd_Success(t *testing.T) {
	//Arrange

	repo := &mockRepo{
		saveFn: func(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
			tx.ID = 1 //симуляция присвоения айди
			return tx, nil
		},
	}
	uc := newTestUsecase(repo)
	tx := domain.Transaction{
		Type:     domain.Income,
		Amount:   5000,
		Category: domain.Salary,
		//Date не передали, проверим что uc заполнит сам
	}

	//Act

	saved, err := uc.Add(context.Background(), tx)

	//Assert

	// t.Fatalf останавливает тест сразу (как return)
	if err != nil {
		t.Fatalf("ожидали нил, получили ошибку: %v", err)
	}
	//там где можем продолжить тест используем errorf
	// он продолжает выполнение — можно увидеть сразу все провалы
	if saved.ID != 1 {
		t.Errorf("ожидали ID = 1, получили ID = %d", saved.ID)
	}
	if saved.Date.IsZero() {
		t.Errorf("дата должна быть заполнена")
	}
}

// тест который предложил клауд
// сверху это все равно проверялось, но напишу
func TestAdd_DateFilledAutomatically(t *testing.T) {
	//Arrange

	repo := &mockRepo{
		saveFn: func(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
			return tx, nil
		},
	}

	uc := newTestUsecase(repo)
	tx := domain.Transaction{
		Type:     domain.Income,
		Amount:   1000,
		Category: domain.Salary,
	}

	//Act

	saved, err := uc.Add(context.Background(), tx)

	//assert

	if err != nil {
		t.Fatalf("ожидали nil, получили ошибку: %v", err)
	}
	if saved.Date.IsZero() {
		t.Errorf("ожидали автоматическую дату, получили нуль")
	}

}

// тестим негативный сценарий
func TestAdd_invalidAmount(t *testing.T) {
	//Arrange

	uc := newTestUsecase((&mockRepo{}))
	tx := domain.Transaction{
		Type:     domain.Income,
		Amount:   0, //невлидно
		Category: domain.Salary,
	}

	//Act

	_, err := uc.Add(context.Background(), tx)

	//Assert
	// нам помогает что я обернул ошибки с помощью плейсхолдера %w
	// и могу их проверять по errors.Is
	if err == nil {
		t.Fatal("ожидали ошибку, получили nil")
	}
	if !errors.Is(err, domain.ErrIvalidAmount) {
		t.Errorf("ожидали ErrInvalidAmount, получили %v", err)
	}
}

func TestAdd_InvalidType(t *testing.T) {
	//Arrange
	uc := newTestUsecase(&mockRepo{})
	tx := domain.Transaction{
		Type:     "invalid",
		Amount:   100,
		Category: domain.Clothes,
	}

	//Act

	_, err := uc.Add(context.Background(), tx)

	//Assert

	if err == nil {
		t.Fatalf("ожидали ошибку, получили нил")
	}
	if !errors.Is(err, domain.ErrInvalidType) {
		t.Errorf("ожидали ИнвалидТайп, получили: %v", err)
	}
}

func TestAdd_InvalidCategory(t *testing.T) {
	//Arrange
	uc := newTestUsecase(&mockRepo{})
	tx := domain.Transaction{
		Type:     domain.Income,
		Amount:   1000,
		Category: "invalid",
	}

	//Act
	_, err := uc.Add(context.Background(), tx)
	//Assert

	if err == nil {
		t.Fatalf("ожидал ошибку, получил нил")

	}
	if !errors.Is(err, domain.ErrInvalidCategory) {
		t.Errorf("получил не тот вид ошибки: %v, ждал ИнвалидКатегори", err)
	}
}

func TestAdd_RepoError(t *testing.T) {
	//Arrange
	repoErr := errors.New("база упала")
	repo := &mockRepo{
		saveFn: func(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
			return domain.Transaction{}, repoErr
		},
	}

	uc := newTestUsecase(repo)
	tx := domain.Transaction{
		Type:     domain.Income,
		Amount:   100,
		Category: domain.Food,
	}

	//Act
	_, err := uc.Add(context.Background(), tx)

	//Assert
	if err == nil {
		t.Fatalf("ожидал шибку, получил нил")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("ожидал ошибку <база упала>, получил %v", err)
	}

}

// тестим получение транзы по айди
func TestGetByID_Success(t *testing.T) {
	//Arrange

	//наш эталонный домеин
	want := domain.Transaction{
		ID:       42,
		Type:     domain.Income,
		Amount:   100,
		Category: domain.Food,
	}

	repo := &mockRepo{
		findByIDFn: func(ctx context.Context, id int64) (domain.Transaction, error) {
			return want, nil
		},
	}
	uc := newTestUsecase(repo)

	//Act

	got, err := uc.Add(context.Background(), want)

	//Assert

	if err != nil {
		t.Fatalf("ожидали нил, получили ошибку %v", err)
	}
	if got.ID != want.ID {
		t.Errorf("ожидали ID = %d, получили %d", want.ID, got.ID)
	}

}

func TestGetById_NotFound(t *testing.T) {
	//Arrange
	repo := &mockRepo{
		findByIDFn: func(ctx context.Context, id int64) (domain.Transaction, error) {
			return domain.Transaction{}, domain.ErrNotFound
		},
	}
	uc := newTestUsecase(repo)

	//Act

	_, err := uc.GetByID(context.Background(), 114124)

	//Assert

	if err == nil {
		t.Fatalf("ожидал ошибку, получил нил")
	}
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ожидал нот фаунд, получил %v", err)
	}
}

func TestGetAll_Success(t *testing.T) {
	//arrange
	want := []domain.Transaction{
		{ID: 1, Type: domain.Income, Amount: 100, Category: domain.Food},
		{ID: 2, Type: domain.Income, Amount: 100, Category: domain.Food},
	}

	repo := &mockRepo{
		findAllFn: func(ctx context.Context) ([]domain.Transaction, error) {
			return want, nil
		},
	}
	uc := newTestUsecase(repo)

	//act
	got, err := uc.GetAll(context.Background())

	//assert
	if err != nil {
		t.Fatalf("ожидали nil, получили: %v", err)
	}

	if len(got) != len(want) {
		t.Errorf("ожидали %d транзакций, получили %d", len(want), len(got))
	}
}

func TestGetAll_RepoError(t *testing.T) {
	repoErr := errors.New("база упала")
	repo := &mockRepo{
		findAllFn: func(ctx context.Context) ([]domain.Transaction, error) {
			return nil, repoErr
		},
	}
	uc := newTestUsecase(repo)

	_, err := uc.GetAll(context.Background())

	if err == nil {
		t.Fatalf("ожидал ошибку, получил нил")
	}
	if !errors.Is(err, repoErr) {
		t.Errorf("ожидали ошибку репо, получили: %v", err)
	}
}

func TestGetByType_Success(t *testing.T) {

	// Arrange
	want := []domain.Transaction{
		{ID: 1, Type: domain.Expense, Amount: 200, Category: domain.Food},
		{ID: 2, Type: domain.Expense, Amount: 150, Category: domain.Transport},
	}
	repo := &mockRepo{
		findByTypeFn: func(ctx context.Context, t domain.Type) ([]domain.Transaction, error) {
			return want, nil
		},
	}
	uc := newTestUsecase(repo)

	// Act
	got, err := uc.GetByType(context.Background(), domain.Expense)

	// Assert
	if err != nil {
		t.Fatalf("ожидали nil, получили: %v", err)
	}
	//проверяем тип у каждой транзы
	for _, tx := range got {
		if tx.Type != domain.Expense {
			t.Errorf("ожидали тип Expence, получили: %v", err)
		}
	}
}

func TestGetByPeriod_Success(t *testing.T) {
	//создаем дату, тесты должны быть детерменированы
	from := time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC)
	to := time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC)

	want := []domain.Transaction{
		{ID: 1, Type: domain.Income, Amount: 5000, Category: domain.Food},
	}
	repo := &mockRepo{
		findByPeriodFn: func(ctx context.Context, f, t time.Time) ([]domain.Transaction, error) {
			return want, nil
		},
	}
	uc := newTestUsecase(repo)

	// Act
	got, err := uc.GetByPeriod(context.Background(), from, to)

	// Assert
	if err != nil {
		t.Fatalf("ожидали nil, получили: %v", err)
	}
	if len(got) != len(want) {
		t.Errorf("ожидали %d транзакций, получили %d", len(want), len(got))
	}
}

// проверяем защиту от невалидного периода.
func TestGetByPeriod_InvalidRange(t *testing.T) {
	// Arrange
	uc := newTestUsecase(&mockRepo{})
	from := time.Now()
	to := from.Add(-24 * time.Hour) // to раньше from — невалидно

	// Act
	_, err := uc.GetByPeriod(context.Background(), from, to)

	// Assert
	if err == nil {
		t.Fatal("ожидали ошибку невалидного периода, получили nil")
	}
}

func TestDelete_Success(t *testing.T) {
	repo := &mockRepo{
		deleteFn: func(ctx context.Context, id int64) error {
			return nil
		},
	}
	uc := newTestUsecase(repo)

	err := uc.Delete(context.Background(), 1)

	if err != nil {
		t.Fatalf("ожидали нил получили ошибку %v:", err)
	}
}

func TestDelete_NotFound(t *testing.T) {
	// Arrange
	repo := &mockRepo{
		deleteFn: func(ctx context.Context, id int64) error {
			return domain.ErrNotFound
		},
	}
	uc := newTestUsecase(repo)

	// Act
	err := uc.Delete(context.Background(), 99)

	// Assert
	if !errors.Is(err, domain.ErrNotFound) {
		t.Errorf("ожидали ErrNotFound, получили: %v", err)
	}
}
