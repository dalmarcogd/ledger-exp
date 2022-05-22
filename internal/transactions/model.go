package transactions

import (
	"time"

	"github.com/dalmarcogd/blockchain-exp/pkg/database"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type transactionModel struct {
	bun.BaseModel `bun:"table:transactions"`

	ID            uuid.UUID `bun:"id,pk"`
	FromAccountID uuid.UUID `bun:"from_account_id,nullzero"`
	ToAccountID   uuid.UUID `bun:"to_account_id,nullzero"`
	Amount        Amount    `bun:"amount"`
	Hash          Hash      `bun:"hash"`
	PrevHash      Hash      `bun:"prev_hash"`
	Description   string    `bun:"description"`
	CreatedAt     time.Time `bun:"created_at,nullzero"`
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

type transactionFilter struct {
	ID             uuid.NullUUID
	FromAccountID  uuid.NullUUID
	ToAccountID    uuid.NullUUID
	Hash           Hash
	PrevHash       Hash
	CreatedAtBegin database.NullTime
	CreatedAtEnd   database.NullTime
}
