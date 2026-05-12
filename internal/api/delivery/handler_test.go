package delivery

import (
	"KolinFinance/internal/domain"
	"KolinFinance/internal/pkg/logger"
	"KolinFinance/internal/ports"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

// конструктор хендлера для тестов
func newTestHandler(txUC *mockTransactionUC, reportUC *mockReportUC) *Handler {
	return &Handler{
		txUC:     txUC,
		reportUC: reportUC,
		log:      logger.New(logger.LevelError),
	}
}

// вспомогательная функция
// эта суета создает запрос и пропускает через роутер,
// возвращая респонс
func sendRequest(t *testing.T, handler *Handler, method, path string, body any) *http.Response {
	//помечает функцию как вспомогательную, не появится в стектрейсе
	t.Helper()

	var reqBody *bytes.Buffer
	if body != nil {
		//маршал превращает строку в то что можно передать по сети
		// (JSON) - слайс байтов
		data, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("ошибка сериализации запроса %v:", err)
		}
		// bytes.NewBuffer(data) и bytes.NewReader(data)
		// создают объект, который содержит внутри твои байты
		// и добавляет к ним методы, необходимые
		// для интерфейса io.Reader (в частности, метод Read()).
		reqBody = bytes.NewBuffer(data)

	} else {
		reqBody = bytes.NewBuffer(nil)
	}

	//тестовый запрос
	req := httptest.NewRequest(method, path, reqBody)
	req.Header.Set("Content-Type", "application/json")

	//записывает ответ хендлера
	recorder := httptest.NewRecorder()

	//прогоняем запрос через роутер
	handler.Router().ServeHTTP(recorder, req)

	return recorder.Result()
}

func decodeJSON(t *testing.T, resp *http.Response, v any) {
	t.Helper()
	//создается декодер и читается поток байт
	//которые записывают в структуру v
	if err := json.NewDecoder(resp.Body).Decode(v); err != nil {
		t.Fatalf("ошибка декодирования ответа: %v", err)
	}
}

// ── POST /transactions ────────────────────────────────────────────────────────
func TestAddTransactions_Success(t *testing.T) {
	//arrange
	want := domain.Transaction{
		ID:       1000,
		Type:     domain.Income,
		Amount:   1000,
		Category: domain.Food,
		Date:     time.Now(),
	}
	txUC := &mockTransactionUC{
		addFn: func(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
			return want, nil
		},
	}
	handler := newTestHandler(txUC, &mockReportUC{})
	body := map[string]any{
		"type":     "income",
		"amount":   1000,
		"category": "salary",
	}

	//act
	resp := sendRequest(t, handler, http.MethodPost, "/transactions/", body)

	//assert
	if resp.StatusCode != http.StatusCreated {
		t.Errorf("ожидали статус 201, получили %d", resp.StatusCode)
	}

	//декодим и проверяем та ли это структура
	var got domain.Transaction
	decodeJSON(t, resp, &got)
	if got.ID != want.ID {
		t.Errorf("ожидали ID=%d, получили ID=%d", want.ID, got.ID)
	}
}

// передаем невалидный json
// ждем ошибку 400
func TestAddTransaction_InvalidBody(t *testing.T) {
	// Arrange
	handler := newTestHandler(&mockTransactionUC{}, &mockReportUC{})

	// создаём запрос с невалидным JSON вручную
	req := httptest.NewRequest(http.MethodPost, "/transactions/", bytes.NewBufferString("not json"))
	req.Header.Set("Content-Type", "application/json")
	//эта штука записывает ответ от сервера
	recorder := httptest.NewRecorder()

	// Act
	handler.Router().ServeHTTP(recorder, req)

	// Assert
	if recorder.Code != http.StatusBadRequest {
		t.Errorf("ожидали статус 400, получили %d", recorder.Code)
	}
}

// проверяем что handler правильно
// обрабатывает ошибки валидации из usecase и возвращает 400
func TestAddTransaction_ValidationError(t *testing.T) {
	// Arrange
	txUC := &mockTransactionUC{
		addFn: func(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
			return domain.Transaction{}, domain.ErrIvalidAmount
		},
	}
	handler := newTestHandler(txUC, &mockReportUC{})
	body := map[string]any{
		"type":     "income",
		"amount":   -100, // usecase вернёт ошибку
		"category": "salary",
	}

	// Act
	resp := sendRequest(t, handler, http.MethodPost, "/transactions/", body)

	// Assert
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("ожидали статус 400, получили %d", resp.StatusCode)
	}
}

func TestAddTransaction_InternalError(t *testing.T) {
	//arrange
	txUc := &mockTransactionUC{
		addFn: func(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {
			return domain.Transaction{}, errors.New("неизвестная ошибка")
		},
	}
	handler := newTestHandler(txUc, &mockReportUC{})
	body := map[string]any{
		"type":     "expense",
		"amount":   1000,
		"category": "food",
	}

	//act
	resp := sendRequest(t, handler, http.MethodPost, "/transactions/", body)

	//assert
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("ожидали статус код 500б получили: %v", resp.StatusCode)
	}
}

// ── GET /transactions ─────────────────────────────────────────────────────────
func TestListTransaction_success(t *testing.T) {
	//arrange
	want := []domain.Transaction{
		{ID: 1, Type: "expense", Amount: 1000, Category: "food"},
		{ID: 2, Type: "expense", Amount: 1000, Category: "food"},
	}
	txUC := &mockTransactionUC{
		getAllFn: func(ctx context.Context) ([]domain.Transaction, error) {
			return want, nil
		},
	}

	handler := newTestHandler(txUC, &mockReportUC{})
	//act
	resp := sendRequest(t, handler, http.MethodGet, "/transactions/", want)

	//assert
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ожидали статус 200, получили %d", resp.StatusCode)
	}

	var got []domain.Transaction
	decodeJSON(t, resp, &got)
	if len(want) != len(got) {
		t.Errorf("количество полученных транзакций: %v, не равно тому что ожидал: %v", len(got), len(want))
	}
}

//ДАЛЬШЕ ПИСАЛ НЕ САМ, ТОЛЬОК ПРОБЕЖАЛСЯ ГЛАЗАМИ

// TestListTransactions_FilterByType — проверяем что query параметр ?type=
// передаётся в usecase, вызывается GetByType а не GetAll.
func TestListTransactions_FilterByType(t *testing.T) {
	// Arrange
	// getByTypeFn заполнен — значит ожидаем что вызовется именно он.
	// getAllFn не заполнен — если вызовется, будет паника.
	called := false
	txUC := &mockTransactionUC{
		getByTypeFn: func(ctx context.Context, t domain.Type) ([]domain.Transaction, error) {
			called = true // фиксируем что метод был вызван
			return []domain.Transaction{}, nil
		},
	}
	handler := newTestHandler(txUC, &mockReportUC{})

	// Act
	resp := sendRequest(t, handler, http.MethodGet, "/transactions/?type=expense", nil)

	// Assert
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ожидали статус 200, получили %d", resp.StatusCode)
	}
	if !called {
		t.Error("ожидали вызов GetByType, но он не был вызван")
	}
}

// ── GET /transactions/:id ─────────────────────────────────────────────────────

func TestGetTransaction_Success(t *testing.T) {
	// Arrange
	want := domain.Transaction{ID: 42, Type: domain.Expense, Amount: 300, Category: domain.Food}
	txUC := &mockTransactionUC{
		getByIDFn: func(ctx context.Context, id int64) (domain.Transaction, error) {
			return want, nil
		},
	}
	handler := newTestHandler(txUC, &mockReportUC{})

	// Act
	resp := sendRequest(t, handler, http.MethodGet, "/transactions/42", nil)

	// Assert
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ожидали статус 200, получили %d", resp.StatusCode)
	}
	var got domain.Transaction
	decodeJSON(t, resp, &got)
	if got.ID != want.ID {
		t.Errorf("ожидали ID=%d, получили ID=%d", want.ID, got.ID)
	}
}

func TestGetTransaction_NotFound(t *testing.T) {
	// Arrange
	txUC := &mockTransactionUC{
		getByIDFn: func(ctx context.Context, id int64) (domain.Transaction, error) {
			return domain.Transaction{}, domain.ErrNotFound
		},
	}
	handler := newTestHandler(txUC, &mockReportUC{})

	// Act
	resp := sendRequest(t, handler, http.MethodGet, "/transactions/99", nil)

	// Assert
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("ожидали статус 404, получили %d", resp.StatusCode)
	}
}

// TestGetTransaction_InvalidID — проверяем что handler отвечает 400
// если id в пути не является числом.
func TestGetTransaction_InvalidID(t *testing.T) {
	// Arrange
	handler := newTestHandler(&mockTransactionUC{}, &mockReportUC{})

	// Act
	// "abc" не парсится в int64 — ожидаем 400
	resp := sendRequest(t, handler, http.MethodGet, "/transactions/abc", nil)

	// Assert
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("ожидали статус 400, получили %d", resp.StatusCode)
	}
}

// ── DELETE /transactions/:id ──────────────────────────────────────────────────

func TestDeleteTransaction_Success(t *testing.T) {
	// Arrange
	txUC := &mockTransactionUC{
		deleteFn: func(ctx context.Context, id int64) error {
			return nil
		},
	}
	handler := newTestHandler(txUC, &mockReportUC{})

	// Act
	resp := sendRequest(t, handler, http.MethodDelete, "/transactions/1", nil)

	// Assert
	// 204 No Content — успешное удаление без тела ответа
	if resp.StatusCode != http.StatusNoContent {
		t.Errorf("ожидали статус 204, получили %d", resp.StatusCode)
	}
}

func TestDeleteTransaction_NotFound(t *testing.T) {
	// Arrange
	txUC := &mockTransactionUC{
		deleteFn: func(ctx context.Context, id int64) error {
			return domain.ErrNotFound
		},
	}
	handler := newTestHandler(txUC, &mockReportUC{})

	// Act
	resp := sendRequest(t, handler, http.MethodDelete, "/transactions/99", nil)

	// Assert
	if resp.StatusCode != http.StatusNotFound {
		t.Errorf("ожидали статус 404, получили %d", resp.StatusCode)
	}
}

// ── GET /report/balance ───────────────────────────────────────────────────────

func TestGetBalance_Success(t *testing.T) {
	// Arrange
	reportUC := &mockReportUC{
		balanceFn: func(ctx context.Context) (float64, error) {
			return 42000.50, nil
		},
	}
	handler := newTestHandler(&mockTransactionUC{}, reportUC)

	// Act
	resp := sendRequest(t, handler, http.MethodGet, "/report/balance", nil)

	// Assert
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ожидали статус 200, получили %d", resp.StatusCode)
	}
	// декодируем {"balance": 42000.50} и проверяем значение
	var got map[string]float64
	decodeJSON(t, resp, &got)
	if got["balance"] != 42000.50 {
		t.Errorf("ожидали баланс 42000.50, получили %v", got["balance"])
	}
}

func TestGetBalance_InternalError(t *testing.T) {
	// Arrange
	reportUC := &mockReportUC{
		balanceFn: func(ctx context.Context) (float64, error) {
			return 0, errors.New("база упала")
		},
	}
	handler := newTestHandler(&mockTransactionUC{}, reportUC)

	// Act
	resp := sendRequest(t, handler, http.MethodGet, "/report/balance", nil)

	// Assert
	if resp.StatusCode != http.StatusInternalServerError {
		t.Errorf("ожидали статус 500, получили %d", resp.StatusCode)
	}
}

// ── GET /report/period ────────────────────────────────────────────────────────

func TestGetReportByPeriod_Success(t *testing.T) {
	// Arrange
	want := ports.Report{
		From:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:      time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		Income:  50000,
		Expense: 15000,
		Balance: 35000,
	}
	reportUC := &mockReportUC{
		reportByPeriodFn: func(ctx context.Context, from, to time.Time) (ports.Report, error) {
			return want, nil
		},
	}
	handler := newTestHandler(&mockTransactionUC{}, reportUC)

	// Act
	// передаём даты через query параметры
	resp := sendRequest(t, handler, http.MethodGet, "/report/period?from=2024-01-01&to=2024-01-31", nil)

	// Assert
	if resp.StatusCode != http.StatusOK {
		t.Errorf("ожидали статус 200, получили %d", resp.StatusCode)
	}
	var got ports.Report
	decodeJSON(t, resp, &got)
	if got.Balance != want.Balance {
		t.Errorf("ожидали баланс %v, получили %v", want.Balance, got.Balance)
	}
}

// TestGetReportByPeriod_InvalidDate — проверяем что handler отвечает 400
// если дата передана в неправильном формате.
func TestGetReportByPeriod_InvalidDate(t *testing.T) {
	// Arrange
	handler := newTestHandler(&mockTransactionUC{}, &mockReportUC{})

	// Act
	// неправильный формат даты — ожидаем 400
	resp := sendRequest(t, handler, http.MethodGet, "/report/period?from=01-01-2024&to=31-01-2024", nil)

	// Assert
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("ожидали статус 400, получили %d", resp.StatusCode)
	}
}

// TestGetReportByPeriod_MissingDate — проверяем что handler отвечает 400
// если query параметры не переданы вообще.
func TestGetReportByPeriod_MissingDate(t *testing.T) {
	// Arrange
	handler := newTestHandler(&mockTransactionUC{}, &mockReportUC{})

	// Act
	resp := sendRequest(t, handler, http.MethodGet, "/report/period", nil)

	// Assert
	if resp.StatusCode != http.StatusBadRequest {
		t.Errorf("ожидали статус 400, получили %d", resp.StatusCode)
	}
}
