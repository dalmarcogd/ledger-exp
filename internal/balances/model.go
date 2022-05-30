package balances

import (
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type accountBalanceModel struct {
	bun.BaseModel `bun:"transactions_balances"`

	AccountID uuid.UUID `bun:"account_id"`
	Balance   float64   `bun:"balance"`
}
