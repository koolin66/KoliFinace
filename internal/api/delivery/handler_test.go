package delivery

import (
	"KolinFinance/internal/domain"
	"KolinFinance/internal/mocks"
	"KolinFinance/internal/pkg/logger"
	"KolinFinance/internal/ports"
	"bytes"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"go.uber.org/mock/gomock"
)

// конструктор хендлера для тестов
func newTestHandler(txUC *mocks.MockTransactionUsecase, reportUC *mocks.MockReportUseCase) *Handler {
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

func TestAddTreansaction_Success(t *testing.T) {
	//arrange
	ctrl := gomock.NewController(t)
	txUC := mocks.NewMockTransactionUsecase(ctrl)

	want := domain.Transaction{ID: 1, Type: domain.Income, Amount: 5000, Category: domain.Salary}

	txUC.EXPECT().Add(gomock.Any(), gomock.Any()).Times(1).Return(want, nil)

	handler := newTestHandler(txUC, mocks.NewMockReportUseCase(ctrl))
	body := map[string]any{"type": "income", "amount": 5000, "category": "salary"}
	//act

	resp := sendRequest(t, handler, http.MethodPost, "/transactions/", body)
	//assert

	assert.Equal(t, http.StatusCreated, resp.StatusCode)
	var got domain.Transaction
	decodeJSON(t, resp, &got)
	assert.Equal(t, want.ID, got.ID)
}

func TestAddTransaction_ValidationError(t *testing.T) {
	//arrange
	ctrl := gomock.NewController(t)
	txUC := mocks.NewMockTransactionUsecase(ctrl)

	txUC.EXPECT().Add(gomock.Any(), gomock.Any()).Times(1).Return(domain.Transaction{}, domain.ErrIvalidAmount)

	handler := newTestHandler(txUC, mocks.NewMockReportUseCase(ctrl))
	body := map[string]any{"type": "income", "amount": -100, "category": "salary"}

	//act
	resp := sendRequest(t, handler, http.MethodPost, "/transactions/", body)

	//assert
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestAddTransaction_InternalError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	txUC := mocks.NewMockTransactionUsecase(ctrl)

	txUC.EXPECT().
		Add(gomock.Any(), gomock.Any()).
		Times(1).
		Return(domain.Transaction{}, errors.New("неизвестная ошибка"))

	handler := newTestHandler(txUC, mocks.NewMockReportUseCase(ctrl))
	body := map[string]any{"type": "income", "amount": 1000, "category": "salary"}

	// Act
	resp := sendRequest(t, handler, http.MethodPost, "/transactions/", body)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// ── GET /transactions ─────────────────────────────────────────────────────────

func TestListTransactions_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	txUC := mocks.NewMockTransactionUsecase(ctrl)

	want := []domain.Transaction{
		{ID: 1, Type: domain.Income, Amount: 5000, Category: domain.Salary},
		{ID: 2, Type: domain.Expense, Amount: 200, Category: domain.Food},
	}
	txUC.EXPECT().
		GetAll(gomock.Any()).
		Times(1).
		Return(want, nil)

	handler := newTestHandler(txUC, mocks.NewMockReportUseCase(ctrl))

	// Act
	resp := sendRequest(t, handler, http.MethodGet, "/transactions/", nil)

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var got []domain.Transaction
	decodeJSON(t, resp, &got)
	assert.Len(t, got, len(want))
}

func TestListTransactions_FilterByType(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	txUC := mocks.NewMockTransactionUsecase(ctrl)

	// GetAll не должен вызываться — только GetByType
	txUC.EXPECT().GetAll(gomock.Any()).Times(0)
	txUC.EXPECT().
		GetByType(gomock.Any(), domain.Expense).
		Times(1).
		Return([]domain.Transaction{}, nil)

	handler := newTestHandler(txUC, mocks.NewMockReportUseCase(ctrl))

	// Act
	resp := sendRequest(t, handler, http.MethodGet, "/transactions/?type=expense", nil)

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// ── GET /transactions/:id ─────────────────────────────────────────────────────

func TestGetTransaction_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	txUC := mocks.NewMockTransactionUsecase(ctrl)

	want := domain.Transaction{ID: 42, Type: domain.Expense, Amount: 300, Category: domain.Food}
	txUC.EXPECT().
		GetByID(gomock.Any(), int64(42)).
		Times(1).
		Return(want, nil)

	handler := newTestHandler(txUC, mocks.NewMockReportUseCase(ctrl))

	// Act
	resp := sendRequest(t, handler, http.MethodGet, "/transactions/42", nil)

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var got domain.Transaction
	decodeJSON(t, resp, &got)
	assert.Equal(t, want.ID, got.ID)
}

func TestGetTransaction_NotFound(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	txUC := mocks.NewMockTransactionUsecase(ctrl)

	txUC.EXPECT().
		GetByID(gomock.Any(), int64(99)).
		Times(1).
		Return(domain.Transaction{}, domain.ErrNotFound)

	handler := newTestHandler(txUC, mocks.NewMockReportUseCase(ctrl))

	// Act
	resp := sendRequest(t, handler, http.MethodGet, "/transactions/99", nil)

	// Assert
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

func TestGetTransaction_InvalidID(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	txUC := mocks.NewMockTransactionUsecase(ctrl)

	// GetByID не должен вызываться — parseID упадёт раньше
	txUC.EXPECT().GetByID(gomock.Any(), gomock.Any()).Times(0)

	handler := newTestHandler(txUC, mocks.NewMockReportUseCase(ctrl))

	// Act
	resp := sendRequest(t, handler, http.MethodGet, "/transactions/abc", nil)

	// Assert
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

// ── DELETE /transactions/:id ──────────────────────────────────────────────────

func TestDeleteTransaction_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	txUC := mocks.NewMockTransactionUsecase(ctrl)

	txUC.EXPECT().
		Delete(gomock.Any(), int64(1)).
		Times(1).
		Return(nil)

	handler := newTestHandler(txUC, mocks.NewMockReportUseCase(ctrl))

	// Act
	resp := sendRequest(t, handler, http.MethodDelete, "/transactions/1", nil)

	// Assert
	assert.Equal(t, http.StatusNoContent, resp.StatusCode)
}

func TestDeleteTransaction_NotFound(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	txUC := mocks.NewMockTransactionUsecase(ctrl)

	txUC.EXPECT().
		Delete(gomock.Any(), int64(99)).
		Times(1).
		Return(domain.ErrNotFound)

	handler := newTestHandler(txUC, mocks.NewMockReportUseCase(ctrl))

	// Act
	resp := sendRequest(t, handler, http.MethodDelete, "/transactions/99", nil)

	// Assert
	assert.Equal(t, http.StatusNotFound, resp.StatusCode)
}

// ── GET /report/balance ───────────────────────────────────────────────────────

func TestGetBalance_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	reportUC := mocks.NewMockReportUseCase(ctrl)

	reportUC.EXPECT().
		Balance(gomock.Any()).
		Times(1).
		Return(42000.50, nil)

	handler := newTestHandler(mocks.NewMockTransactionUsecase(ctrl), reportUC)

	// Act
	resp := sendRequest(t, handler, http.MethodGet, "/report/balance", nil)

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var got map[string]float64
	decodeJSON(t, resp, &got)
	assert.Equal(t, 42000.50, got["balance"])
}

func TestGetBalance_InternalError(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	reportUC := mocks.NewMockReportUseCase(ctrl)

	reportUC.EXPECT().
		Balance(gomock.Any()).
		Times(1).
		Return(0.0, errors.New("база упала"))

	handler := newTestHandler(mocks.NewMockTransactionUsecase(ctrl), reportUC)

	// Act
	resp := sendRequest(t, handler, http.MethodGet, "/report/balance", nil)

	// Assert
	assert.Equal(t, http.StatusInternalServerError, resp.StatusCode)
}

// ── GET /report/period ────────────────────────────────────────────────────────

func TestGetReportByPeriod_Success(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	reportUC := mocks.NewMockReportUseCase(ctrl)

	want := ports.Report{
		From:    time.Date(2024, 1, 1, 0, 0, 0, 0, time.UTC),
		To:      time.Date(2024, 1, 31, 0, 0, 0, 0, time.UTC),
		Income:  50000,
		Expense: 15000,
		Balance: 35000,
	}

	// gomock.Any() для дат — точное совпадение time.Time нестабильно
	// из-за часовых поясов при парсинге из строки
	reportUC.EXPECT().
		ReportByPeriod(gomock.Any(), gomock.Any(), gomock.Any()).
		Times(1).
		Return(want, nil)

	handler := newTestHandler(mocks.NewMockTransactionUsecase(ctrl), reportUC)

	// Act
	resp := sendRequest(t, handler, http.MethodGet, "/report/period?from=2024-01-01&to=2024-01-31", nil)

	// Assert
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	var got ports.Report
	decodeJSON(t, resp, &got)
	assert.Equal(t, want.Balance, got.Balance)
}

func TestGetReportByPeriod_InvalidDate(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	reportUC := mocks.NewMockReportUseCase(ctrl)

	// ReportByPeriod не должен вызываться — parseDate упадёт раньше
	reportUC.EXPECT().ReportByPeriod(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	handler := newTestHandler(mocks.NewMockTransactionUsecase(ctrl), reportUC)

	// Act
	resp := sendRequest(t, handler, http.MethodGet, "/report/period?from=01-01-2024&to=31-01-2024", nil)

	// Assert
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}

func TestGetReportByPeriod_MissingDate(t *testing.T) {
	// Arrange
	ctrl := gomock.NewController(t)
	reportUC := mocks.NewMockReportUseCase(ctrl)
	reportUC.EXPECT().ReportByPeriod(gomock.Any(), gomock.Any(), gomock.Any()).Times(0)

	handler := newTestHandler(mocks.NewMockTransactionUsecase(ctrl), reportUC)

	// Act
	resp := sendRequest(t, handler, http.MethodGet, "/report/period", nil)

	// Assert
	assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
}
