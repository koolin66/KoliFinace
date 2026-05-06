//создание таблицы для доменчиков

// айди бигсериал (до 9 квинтильонов)
// варчар(10) - не больше 10 символов
// нумерик(12,2) 12 чисел, два из которых после запятой - макс (9,999,999,999.99)

package pg

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
)

const createOperationsTable = `
CREATE TABLE IF NOT EXISTS transactions (
    id         BIGSERIAL PRIMARY KEY,
    type       VARCHAR(10)     NOT NULL,
    amount     NUMERIC(12, 2)  NOT NULL,
    category   VARCHAR(50)     NOT NULL,
    date       TIMESTAMPTZ     NOT NULL,
    note       TEXT            NOT NULL DEFAULT ''
);`

// создаёт таблицу, если её ещё нет
func Migrate(ctx context.Context, db *pgxpool.Pool) error {
	_, err := db.Exec(ctx, createOperationsTable)
	return err
}
