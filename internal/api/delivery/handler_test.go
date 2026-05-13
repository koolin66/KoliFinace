package delivery

import (
	"KolinFinance/internal/mocks"
	"KolinFinance/internal/pkg/logger"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
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
