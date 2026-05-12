package delivery

import (
	"KolinFinance/internal/domain"
	"KolinFinance/internal/pkg/logger"
	"KolinFinance/internal/ports"
	"errors"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

type ServerConfig struct {
	Host string `env:"HOST" default:"0.0.0.0"`
	Port string `env:"PORT" default:"8080"`
}

type Handler struct {
	txUC     ports.TransactionUsecase
	reportUC ports.ReportUseCase
	log      *logger.Logger
}

func NewHandler(txUC ports.TransactionUsecase, reportUC ports.ReportUseCase, log *logger.Logger) *Handler {
	return &Handler{txUC: txUC, reportUC: reportUC, log: log}
}

//-----------------------------------------------------------------------
//Что конкретно gin берёт на себя по сравнению с тем что ты уже написал руками:
// роутинг по методу — не нужен switch r.Method
// URL параметры — не нужен strings.Split(r.URL.Path, "/")
// c.ShouldBindJSON — валидация через теги вместо голого json.Decode
// логгер уже встроен — не нужен свой middleware для базового логирования

// тот самый роутер на http/net
func (h *Handler) Router() *gin.Engine {
	//создается роутер
	r := gin.New()
	//ловит панику вместо краша, возвращает 500
	r.Use(gin.Recovery())
	//логгирует каждый запрос
	r.Use(h.loggerMiddleware())

	//создаем группы
	//Параметр :id: Gin извлекает его и кладёт в c.Param("id").
	tx := r.Group("/transactions")
	{
		tx.POST("/", h.addTransaction)
		tx.GET("/", h.listTransactions)
		tx.GET("/:id", h.getTransaction)
		tx.DELETE("/:id", h.deleteTransaction)
	}

	report := r.Group("/report")
	{
		report.GET("/balance", h.getBalance)
		report.GET("/period", h.getReportByPeriod)
	}

	return r
}

// -----------------------------------------------------------------------
// короче, gin хранит функции в слайсе, и с помощью функции c.Next,
//проходит по нему и вызывает всех по очереди

func (h *Handler) loggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Next() //здесь вызывается след функция в цепочке
		h.log.Info("request",
			"method", c.Request.Method,
			"path", c.Request.URL.Path,
			"status", c.Writer.Status())

	}
}

// Начало: index = -1

// 1. Gin вызывает loggerMiddleware:
//    - index = 0 (указывает на loggerMiddleware)
//    - выполняется код ДО c.Next()

// 2. loggerMiddleware вызывает c.Next():
//    - index++ → index = 1
//    - вызывает addTransaction (следующий в списке)

// 3. addTransaction выполняется:
//    - устанавливает статус 201
//    - пишет JSON
//    - завершается

// 4. Возврат в loggerMiddleware:
//    - продолжает выполнение ПОСЛЕ c.Next()
//    - логирует статус 201

// 5. Завершение

// ---------------------DTO--------------------------------------------------
// тут уже подхендлеры (методы)

// структура для приема запроса с от клиента, формат должен быть такой
// указатель на time нужен чтобы тут мог быть нил и можно было
// отличить переданое время от непереданого
type addTransactionRequest struct {
	Type     domain.Type     `json:"type"`
	Amount   int64           `json:"amount"`
	Category domain.Category `json:"category"`
	Date     *time.Time      `json:"date"`
	Note     string          `json:"note"`
}

//-----------------------------------------------------------------------

// метод POST
func (h *Handler) addTransaction(c *gin.Context) {
	var req addTransactionRequest

	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"невалидный запрос": err.Error()})
		return //H это такое сокращение для мапы, чтобы тут не писать много
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
	saved, err := h.txUC.Add(c.Request.Context(), tx)
	if err != nil {
		if errors.Is(err, domain.ErrIvalidAmount) ||
			errors.Is(err, domain.ErrInvalidType) ||
			errors.Is(err, domain.ErrInvalidCategory) {
			c.JSON(http.StatusBadRequest, gin.H{"невалидный запрос": err.Error()})
			return
		} //тут не err, потому что вернуть нужно строку
		c.JSON(http.StatusInternalServerError, gin.H{"ошибка добавления транзакции": err.Error()})
	}
	c.JSON(http.StatusCreated, saved)

}

// GET
func (h *Handler) listTransactions(c *gin.Context) {
	txType := domain.Type(c.Query("type"))

	var (
		txs []domain.Transaction
		err error
	)

	if txType != "" {
		txs, err = h.txUC.GetByType(c.Request.Context(), txType)
	} else {
		txs, err = h.txUC.GetAll(c.Request.Context())
	}

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка получения транзакции"})
		return
	}

	c.JSON(http.StatusOK, txs)

}

// GET
func (h *Handler) getTransaction(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный запрос"})
		return

	}

	tx, err := h.txUC.GetByID(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "ошибка получения транзы по айди"})
		return

	}

	c.JSON(http.StatusOK, tx)

}

// DELETE
func (h *Handler) deleteTransaction(c *gin.Context) {
	id, err := parseID(c.Param("id"))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидный айди"})
		return
	}

	err = h.txUC.Delete(c.Request.Context(), id)
	if err != nil {
		if errors.Is(err, domain.ErrNotFound) {
			c.JSON(http.StatusNotFound, gin.H{"error": "транзакция не найдена"})
			return
		} else {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка удаления транзакции"})
			return
		}
	}

	c.Status(http.StatusNoContent)
}

//-----------------------------------------------------------------------

// report хендлеры, оба GET
func (h *Handler) getBalance(c *gin.Context) {
	balance, err := h.reportUC.Balance(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка получения баланса"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"balance": balance})

}

// тут нужно будет задавать  query параметры для того чтобы получить данные
// http://localhost:8080/report/period?from=2024-01-01&to=2024-12-31 - такого плана
func (h *Handler) getReportByPeriod(c *gin.Context) {
	from, err := parseDate(c.Query("from")) // тут ищет параметр from
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидная дата ОТ, используй ГГГГ-ММ-ДД"})
		return
	}

	to, err := parseDate(c.Query("to")) //тут параметр to
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "невалидная дата ДО, используй ГГГГ-ММ-ДД"})
		return
	}

	report, err := h.reportUC.ReportByPeriod(c.Request.Context(), from, to)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "ошибка получения отчета за период"})
		return
	}

	c.JSON(http.StatusOK, report)
}

//-----------------------------------------------------------------------
//вспомогательные функции

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
