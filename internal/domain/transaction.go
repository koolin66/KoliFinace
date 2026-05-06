package domain

import (
	"errors"
	"time"
)

type Type string

const (
	Income  Type = "income"
	Expense Type = "expense"
)

type Category string

const (
	//доходы
	Salary Category = "salary"
	Gift   Category = "gift"

	//расходы
	Food      Category = "food"
	Transport Category = "transport"
	Health    Category = "health"
	Clothes   Category = "clothes"
	Other     Category = "other"
)

// основная сущность приложения
type Transaction struct {
	ID       int64
	Type     Type
	Amount   int64
	Category Category
	Date     time.Time
	Note     string
}

// создадим возможные ошибки для транзакций
var (
	ErrIvalidAmount = errors.New("Сумма должна быть больше 0")
	ErrInvalidType  = errors.New(
		"Тип транзакции может быть только income или expense")
	ErrInvalidCategory = errors.New("Неизвестная категория")
	ErrNotFound        = errors.New("Транзакция не найдена")
)

//сделаем валидацию
//Validate() живёт на самой сущности,
// а не в usecase — это бизнес-правило,
// не зависящее от хранилища.
// Ошибки объявлены здесь же, usecase будет их возвращать
// наверх без оборачивания в новые типы.

func (t Transaction) Validate() error {
	if t.Amount <= 0 {
		return ErrIvalidAmount
	}
	if t.Type != Income && t.Type != Expense {
		return ErrInvalidType
	}
	if !isValidCategory(t.Category) {
		return ErrInvalidCategory
	}
	return nil
}

// напишем отдельную функцию валидации для Category
// уж очень много перечислять.
func isValidCategory(c Category) bool {
	switch c {
	case Salary, Gift, Food, Transport, Health, Other:
		return true
	}
	return false
}
