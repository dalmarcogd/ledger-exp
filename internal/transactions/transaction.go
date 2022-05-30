package transactions

import (
	"github.com/google/uuid"
)

type Transaction struct {
	ID          uuid.UUID
	From        uuid.UUID
	To          uuid.UUID
	Amount      float64
	Description string
}

func newTransaction(model transactionModel) Transaction {
	return Transaction{
		ID:          model.ID,
		From:        model.FromAccountID,
		To:          model.ToAccountID,
		Amount:      model.Amount,
		Description: model.Description,
	}
}
