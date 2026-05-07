package http

import (
	"KolinFinance/internal/domain"
	"KolinFinance/internal/pkg/logger"
	"KolinFinance/internal/ports"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"time"
)

type Handler struct {
	txUC     ports.TransactionUsecase
	reportUC ports.ReportUseCase
	log      *logger.Logger
}

func NewHandler(txUC ports.TransactionUsecase, reportUC ports.ReportUseCase, log *logger.Logger) *Handler {
	return &Handler{txUC: txUC, reportUC: reportUC, log: log}
}

//-----------------------------------------------------------------------

// тот самый роутер на http/net
func (h *Handler) Router() http.Handler {
	// 	создается роутер (мультиплексор (multiplexer - поэтому mux))
	// он хранит таблицу с подключаниями
	mux := http.NewServeMux()

	//	это регистрация маршрута в роутере
	//	logger- обертка для логгирования
	//	http.HandleFunc оборачивает функцию чтобы она реализовывала интерфейс
	//  		http.Handler, ведь мукс принимает только его, у котороего есть метод ServeHTTP

	mux.Handle("/transactions", h.loggerMiddleware(http.HandlerFunc(h.handleTransactions)))
	mux.Handle("/transactions/", h.loggerMiddleware(http.HandlerFunc(h.handleTransactionsByID)))
	mux.Handle("/report/balance", h.loggerMiddleware(http.HandlerFunc(h.getBalance)))
	mux.Handle("/report/period", h.loggerMiddleware(http.HandlerFunc(h.getReportByPeriod)))

	// передаем настроеный роутер туда где будем запускать сервак
	return mux
}

// -----------------------------------------------------------------------
// middlewarчик , в нем логгер
// next - это следующий обработчик в запросе
func (h *Handler) loggerMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		//тут он вызывает нашу следующую функцию хендлер
		next.ServeHTTP(w, r)
		h.log.Info("request",
			"method", r.Method,
			"path", r.URL.Path)
	})

}

//последовательность при вызове функции логгера
// 1. Приходит запрос на /transactions
// 2. mux вызывает loggerMiddleware (первым в цепочке)
// 3. loggerMiddleware логирует? НЕТ, сначала вызывает next!
// 4. next — это (основной хендлер)
//    - Он обрабатывает запрос, формирует ответ
// 5. Управление возвращается в loggerMiddleware
// 6. Логируем: метод, путь, статус (в этом примере без статуса)
// 7. Ответ отправляется клиенту

//-----------------------------------------------------------------------

// это хендлер для методов add и list (добвавить и достать все)
// они там дергаются при смене методов POST, GET
func (h *Handler) handleTransactions(w http.ResponseWriter, r *http.Request) {
	switch t.Method {
	case http.MethodPost:
		h.addTransaction(w, r)
	case http.MethodGet:
		h.getTransaction(w, r)
	default:
		h.writeError()
	}
}

// этот для методов get и delete (все что по айдишнику)
// также зависит от выбраного метода
func (h *Handler) handleTransactionsByID(w http.ResponseWriter, r *http.Request) {

}

// -----------------------------------------------------------------------
// тут уже подхендлеры (методы)

// структура для приема запроса с от клиента, формат должен быть такой
// указатель на time нужен чтобы тут мог быть нил и можно было
// отличить переданое время от непереданого
type addTransactionRequest struct {
	Type     domain.Type     `json: "type`
	Amount   float64         `json: "amount"`
	Category domain.Category `json: "category"`
	Date     *time.Time      `json: "date"`
	Note     string          `json: "note"`
}

// метод POST
func (h *Handler) addTransaction(w http.ResponseWriter, r *http.Request) {
	var req addTransactionRequest

	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		h.writeError(w, http.StatusBadRequest, "invalid request body")
		return
	}

	tx := domain.Transaction{
		Type:     req.Type,
		Amount:   int64(req.Amount),
		Category: req.Category,
		Note:     req.Note,
	}
	if req.Date != nil {
		tx.Date = *req.Date
	}
	//передается контекст из http.Request
	saved, err := h.txUC.Add(r.Context(), tx)
	if err != nil {
		if errors.Is(err, domain.ErrIvalidAmount) ||
			errors.Is(err, domain.ErrInvalidType) ||
			errors.Is(err, domain.ErrInvalidCategory) {
			h.writeError(w, http.StatusInternalServerError, "ошибка добавления транзакции")
			return
		}
	}
	h.writeJSON(w, http.StatusCreated, saved)

}

// GET
func (h *Handler) listTransactions(w http.ResponseWriter, r *http.Request) {
	txType := domain.Type(r.URL.Query().Get("type"))

	var (
		txs []domain.Transaction
		err error
	)

	if txType != "" {
		txs, err = h.txUC.GetByType(r.Context(), txType)
	} else {
		txs, err = h.txUC.GetAll(r.Context())
	}

	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "ошибка получения тразакций")
		return
	}

	h.writeJSON(w, http.StatusOK, txs)
}

// GET
func (h *Handler) getTransaction(w http.ResponseWriter, r *http.Request, rawId string) {
	id, err := parseID(rawId)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "невалидный id")
		return
	}

	tx, err := h.txUC.GetByID(r.Context(), id)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "ошибка получения транзакции по айди")
		return
	}

	h.writeJSON(w, http.StatusOK, tx)
}

// DELETE
func (h *Handler) deleteTransaction(w http.ResponseWriter, r *http.Request, rawID string) {
	id, err := parseID(rawID)
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "невалидный id")
	}

	err = h.txUC.Delete(r.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			h.writeError(w, http.StatusNotFound, "транза не найдена")
			return
		} else {
			h.writeError(w, http.StatusInternalServerError, "ошибка удаления транзакции")
			return
		}
	}

	w.WriteHeader(http.StatusNoContent)
}

//-----------------------------------------------------------------------

// report хендлеры, оба GET
func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "неправильный метод")
		return
	}

	balance, err := h.reportUC.Balance(r.Context())
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "ошибка получения баланса")
		return
	}

	h.writeJSON(w, http.StatusOK, map[string]float64{"balance": balance})
}

func (h *Handler) getReportByPeriod(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		h.writeError(w, http.StatusMethodNotAllowed, "неподходящий метод")
		return
	}

	from, err := parseDate(r.URL.Query().Get("from"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "невалидная дата ОТ, используй ГГГГ-ММ-ДД")
		return
	}

	to, err := parseDate(r.URL.Query().Get("to"))
	if err != nil {
		h.writeError(w, http.StatusBadRequest, "невалидная дата ДО, используй ГГГГ-ММ-ДД")
		return
	}

	report, err := h.reportUC.ReportByPeriod(r.Context(), from, to)
	if err != nil {
		h.writeError(w, http.StatusInternalServerError, "ошибка получения отчета за период")
		return
	}

	h.writeJSON(w, http.StatusOK, report)
}

//-----------------------------------------------------------------------
//вспомогательные функции

//эта кодирует ответ клиенту в джейсик

func (h *Handler) writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	//эта строчка VVVVV создает кодировщик, encode кодирует данные ответа и отправляет
	if err := json.NewEncoder(w).Encode(v); err != nil {
		h.log.Error("ошибка кодировки ответа в джейсОн", "err", err)
	}
}

// тут мапа потому что она парсится на (еррор: мсг) для json ответа
func (h *Handler) writeError(w http.ResponseWriter, status int, msg string) {
	h.writeJSON(w, status, map[string]string{"error": msg})
}

// преобразует строку в число
func parseID(s string) (int64, error) {
	return strconv.ParseInt(s, 10, 64)
}

//                ↑   ↑   ↑
//                |   |   └── 64 бита (int64)
//                |   └────── основание системы счисления (10 = десятичная)
//                └────────── строка для преобразования

// преобразует строку в дату time.Time
func parseDate(s string) (time.Time, error) {
	return time.Parse("2006-01-02", s)
} //					это шаблон
//		ОН ВСЕГДА должен иметь дату 15:04 2 января 2006 года
