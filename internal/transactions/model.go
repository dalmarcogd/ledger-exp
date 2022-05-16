package transactions

import (
	"time"

	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type transactionModel struct {
	bun.BaseModel `bun:"table:transactions"`

	ID            uuid.UUID `bun:"id,pk"`
	FromAccountID uuid.UUID `bun:"from_account_id"`
	ToAccountID   uuid.UUID `bun:"to_account_id"`
	Amount        Amount    `bun:"amount"`
	Hash          Hash      `bun:"hash"`
	PrevHash      Hash      `bun:"prev_hash"`
	CreatedAt     time.Time `bun:"created_at"`
}

func newTransactionModel(tx Transaction) transactionModel {
	return transactionModel{
		ID:            tx.ID,
		FromAccountID: tx.From.ID,
		ToAccountID:   tx.To.ID,
		Amount:        tx.Amount,
		Hash:          tx.Hash,
		PrevHash:      tx.PrevHash,
		CreatedAt:     time.Now().UTC(),
	}
}
