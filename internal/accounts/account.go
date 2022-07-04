package accounts

import "github.com/google/uuid"

type Account struct {
	ID             uuid.UUID
	Name           string
	Agency         string
	Number         string
	DocumentNumber string
	HolderID       uuid.UUID
	Status         Status
}

func newAccount(model accountModel) Account {
	return Account{
		ID:             model.ID,
		Name:           model.Name,
		Agency:         model.Agency,
		Number:         model.Number,
		HolderID:       model.HolderID,
		DocumentNumber: model.HolderDocumentNumber,
		Status:         model.Status,
	}
}

type ListFilter struct {
	Sort           int
	Page           int
	Size           int
	DocumentNumber string
}
