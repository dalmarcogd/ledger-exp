package accounts

import "github.com/google/uuid"

type Account struct {
	ID   uuid.UUID
	Name string
}

func newAccount(model accountModel) Account {
	return Account{
		ID:   model.ID,
		Name: model.Name,
	}
}
