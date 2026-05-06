package http

import (
	"KolinFinance/internal/pkg/logger"
	"KolinFinance/internal/ports"
	"net/http"
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
	mux := http.NewServeMux()

	mux.Handle("/transactions", h.loggerMiddleware(http.HandlerFunc(h.handleTransactions)))
}

// -----------------------------------------------------------------------
// middlewarчик , в нем логгер
func (h *Handler) loggerMiddleware(next http.Handler) http.Handler {

}

//-----------------------------------------------------------------------

// это хендлер для методов add и list (добвавить и достать все)
// они там дергаются при смене методов POST, GET
func (h *Handler) handleTransactions(w http.ResponseWriter, r *http.Request) {

}

// этот для методов get и delete (все что по айдишнику)
// также зависит от выбраного метода
func (h *Handler) handleTransactionsByID(w http.ResponseWriter, r *http.Request) {

}

// -----------------------------------------------------------------------
// тут уже подхендлеры (методы)
// метод POST
func (h *Handler) addTransaction(w http.ResponseWriter, r *http.Request) {

}

// GET
func (h *Handler) listTransactions(w http.ResponseWriter, r *http.Request) {

}

// GET
func (h *Handler) getTransactions(w http.ResponseWriter, r *http.Request) {

}

// DELETE
func (h *Handler) deleteTransaction(w http.ResponseWriter, r *http.Request) {

}

//-----------------------------------------------------------------------

// report хендлеры
func (h *Handler) getBalance(w http.ResponseWriter, r *http.Request) {

}

func (h *Handler) getReportByPeriod(w http.ResponseWriter, r *http.Request) {

}
