package holders

import (
	"time"

	"github.com/dalmarcogd/ledger-exp/pkg/database"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type HolderModel struct {
	bun.BaseModel `bun:"table:holders"`

	ID             uuid.UUID `bun:"id,pk"`
	Name           string    `bun:"name"`
	DocumentNumber string    `bun:"document_number"`
	CreatedAt      time.Time `bun:"created_at,notnull"`
	UpdatedAt      time.Time `bun:"updated_at,nullzero"`
}

func newHolderModel(h Holder) HolderModel {
	return HolderModel{
		ID:             h.ID,
		Name:           h.Name,
		DocumentNumber: h.DocumentNumber,
	}
}

type HolderFilter struct {
	ID             uuid.NullUUID
	Name           string
	DocumentNumber string
	CreatedAtBegin database.NullTime
	CreatedAtEnd   database.NullTime
	UpdatedAtBegin database.NullTime
	UpdatedAtEnd   database.NullTime
}

type ListFilter struct {
	Sort           int
	Page           int
	Size           int
	DocumentNumber string
}
