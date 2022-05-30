package accounts

import (
	"time"

	"github.com/dalmarcogd/ledger-exp/pkg/database"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type accountModel struct {
	bun.BaseModel `bun:"table:accounts"`

	ID        uuid.UUID `bun:"id,pk"`
	Name      string    `bun:"name"`
	CreatedAt time.Time `bun:"created_at,notnull"`
	UpdatedAt time.Time `bun:"updated_at,nullzero"`
}

func newAccountModel(acc Account) accountModel {
	return accountModel{
		ID:   acc.ID,
		Name: acc.Name,
	}
}

type accountFilter struct {
	ID             uuid.NullUUID
	Name           string
	CreatedAtBegin database.NullTime
	CreatedAtEnd   database.NullTime
	UpdatedAtBegin database.NullTime
	UpdatedAtEnd   database.NullTime
}
