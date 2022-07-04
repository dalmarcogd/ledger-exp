package transactions

import (
	"github.com/google/uuid"
)

type TransactionType string

var (
	CreditTransaction TransactionType = "CREDIT"
	DebitTransaction  TransactionType = "DEBIT"
	P2PTransaction    TransactionType = "P2P"
)

type Transaction struct {
	ID          uuid.UUID
	From        uuid.UUID
	To          uuid.UUID
	Type        TransactionType
	Amount      float64
	Description string
}

func newTransaction(model transactionModel) Transaction {
	return Transaction{
		ID:          model.ID,
		From:        model.FromAccountID,
		To:          model.ToAccountID,
		Type:        model.Type,
		Amount:      model.Amount,
		Description: model.Description,
	}
}
