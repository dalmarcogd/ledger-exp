package transactions

import (
	"time"

	"github.com/dalmarcogd/ledger-exp/pkg/database"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type transactionModel struct {
	bun.BaseModel `bun:"table:transactions"`

	ID            uuid.UUID `bun:"id,pk"`
	FromAccountID uuid.UUID `bun:"from_account_id,nullzero"`
	ToAccountID   uuid.UUID `bun:"to_account_id,nullzero"`
	Amount        float64   `bun:"amount"`
	Description   string    `bun:"description"`
	CreatedAt     time.Time `bun:"created_at,notnull"`
}

func newTransactionModel(tx Transaction) transactionModel {
	return transactionModel{
		ID:            uuid.New(),
		FromAccountID: tx.From,
		ToAccountID:   tx.To,
		Amount:        tx.Amount,
		Description:   tx.Description,
		CreatedAt:     time.Now().UTC(),
	}
}

type transactionFilter struct {
	ID             uuid.NullUUID
	FromAccountID  uuid.NullUUID
	ToAccountID    uuid.NullUUID
	CreatedAtBegin database.NullTime
	CreatedAtEnd   database.NullTime
}
