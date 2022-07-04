package holders

import "github.com/google/uuid"

type Holder struct {
	ID             uuid.UUID
	Name           string
	DocumentNumber string
}

func newHolder(model HolderModel) Holder {
	return Holder{
		ID:             model.ID,
		Name:           model.Name,
		DocumentNumber: model.DocumentNumber,
	}
}
