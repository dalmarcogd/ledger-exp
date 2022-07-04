package transactionsh

type createdTransaction struct {
	ID          string  `json:"id"`
	From        string  `json:"from_account_id,omitempty"`
	To          string  `json:"to_account_id,omitempty"`
	Type        string  `json:"type"`
	Amount      float64 `json:"amount"`
	Description string  `json:"description"`
}
