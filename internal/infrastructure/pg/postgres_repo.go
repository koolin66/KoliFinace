//методы КРУД для транзакций

package pg

import (
	"KolinFinance/internal/domain"
	"context"
	"errors"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

// оберточка над пулом соединений для посгрес
// библиотеку pgx мы можем менять на что то другое с такими же методами
// без сложных операций
type postgresRepo struct {
	pool *pgxpool.Pool
}

// конструктор создает экземпляр репо с прокинутым пулом
func NewPostgresRepo(pool *pgxpool.Pool) *postgresRepo {
	return &postgresRepo{pool: pool}
}

// метод сохранения транзы в бд
// возвращает транзу чтобы оттуда можно было достать id, ведь он присвоился
func (r *postgresRepo) Save(ctx context.Context, tx domain.Transaction) (domain.Transaction, error) {

	//запихиваем транзу в бд и возвращаем id
	query := `
		INSERT INTO transactions (type, amount, category, date, note)
        VALUES ($1, $2, $3, $4, $5)
        RETURNING id`

	//тут суета выполняет запрос, возварщает Id и кладет его в структуру транзы
	err := r.pool.QueryRow(ctx, query,
		tx.Type,
		tx.Amount,
		tx.Category,
		tx.Date,
		tx.Note).Scan(&tx.ID)
	if err != nil {
		return domain.Transaction{}, err
	}
	return tx, nil
}

func (r *postgresRepo) FindById(ctx context.Context, id int64) (domain.Transaction, error) {
	query := `
        SELECT id, type, amount, category, date, note
        FROM transactions
        WHERE id = $1`

	//тут .Scan вынесем в отельную функцию scanTransaction
	//чтобы не писать кучу всего в этой
	tx, err := scanTransaction(r.pool.QueryRow(ctx, query, id))
	if errors.Is(err, pgx.ErrNoRows) {
		return domain.Transaction{}, domain.ErrNotFound
	}
	return tx, err
}

// тоже вспомогательная функция для транзатионССС!
func (r *postgresRepo) FindAll(ctx context.Context) ([]domain.Transaction, error) {
	query := `
        SELECT id, type, amount, category, date, note
        FROM transactions
        ORDER BY date DESC`

	// также вспомогательная функция в нее прокидываем Query, a не QueryRow
	return scanTransactions(r.pool.Query(ctx, query))
}

// особо нечего писать, все понятно
func (r *postgresRepo) FindByType(ctx context.Context, t domain.Type) ([]domain.Transaction, error) {
	query := `
        SELECT id, type, amount, category, date, note
        FROM transactions
        WHERE type = $1
        ORDER BY date DESC`
	// тут суета Query прокидывает t в запрос и получает ответ в виде
	// нескольких строк, дальше все знаешь
	return scanTransactions(r.pool.Query(ctx, query, t))
}

// все тоже самое что и выше, только запрос сложнее
func (r *postgresRepo) FindByPeriod(ctx context.Context, from, to time.Time) ([]domain.Transaction, error) {
	query := `
        SELECT id, type, amount, category, date, note
        FROM transactions
        WHERE date >= $1 AND date <= $2
        ORDER BY date DESC`

	return scanTransactions(r.pool.Query(ctx, query, from, to))
}

func (r *postgresRepo) Delete(ctx context.Context, id int64) error {
	query := `DELETE FROM transactions WHERE id = $1`

	//Exex используется для запросов которые ничего не возвращают
	result, err := r.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}
	// результ возвращает количество затронутых строк (в бд),
	// типа если ничего не потрогал, то ничего и не сделал - ошибка

	if result.RowsAffected() == 0 {
		return domain.ErrNotFound
	}
	return nil
}

// -------------------------------------------------------------------
// функции помощники
// -------------------------------------------------------------------
func scanTransaction(row pgx.Row) (domain.Transaction, error) {
	var tx domain.Transaction
	err := row.Scan(
		&tx.ID,
		&tx.Type,
		&tx.Amount,
		&tx.Category,
		&tx.Date,
		&tx.Note,
	)
	return tx, err
}

func scanTransactions(rows pgx.Rows, err error) ([]domain.Transaction, error) {
	if err != nil {
		return nil, err
	}
	//дефер гарантирует закрытие соединения
	// его нужно закрыть после чтения, потому что само оно не закроется
	// как в случае с QueryRow где всего одна строка
	defer rows.Close()

	var txs []domain.Transaction
	//цикл по всем строкам в ответе с бд
	//порядок в колонказ должен совпадать с СЕЛЕКТ
	for rows.Next() {
		var tx domain.Transaction
		if err := rows.Scan(
			&tx.ID,
			&tx.Type,
			&tx.Amount,
			&tx.Category,
			&tx.Date,
			&tx.Note,
		); err != nil {
			return nil, err
		}
		txs = append(txs, tx)
	}
	//ретернит слайс и проверяет были ли ошибки во время итерации
	return txs, rows.Err()

}
