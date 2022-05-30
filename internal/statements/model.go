package statements

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type (
	statementModel struct {
		bun.BaseModel `bun:"table:transactions,alias:trx"`

		ID              uuid.UUID `bun:"id,pk"`
		FromAccountID   uuid.UUID `bun:"from_account_id"`
		FromAccountName string    `bun:"from_account_name"`
		ToAccountID     uuid.UUID `bun:"to_account_id"`
		ToAccountName   string    `bun:"to_account_name"`
		Amount          float64   `bun:"amount"`
		Description     string    `bun:"description"`
		CreatedAt       time.Time `bun:"created_at"`
	}

	statementFilter struct {
		Page      int
		Size      int
		Sort      int
		AccountID uuid.UUID
	}
)
