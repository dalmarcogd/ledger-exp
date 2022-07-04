package accounts

import (
	"time"

	"github.com/dalmarcogd/ledger-exp/pkg/database"
	"github.com/google/uuid"
	"github.com/uptrace/bun"
)

type Status string

const (
	ActiveStatus  Status = "ACTIVE"
	BlockedStatus Status = "BLOCKED"
	ClosedStatus  Status = "CLOSED"
)

const (
	AccountNumberVariants = "0123456789"
	AccountNumberSize     = 6
)

type accountModel struct {
	bun.BaseModel `bun:"table:accounts"`

	ID                   uuid.UUID `bun:"id,pk"`
	Name                 string    `bun:"name"`
	Agency               string    `bun:"agency"`
	Number               string    `bun:"number"`
	HolderID             uuid.UUID `bun:"holder_id"`
	HolderDocumentNumber string    `bun:"holder_document_number,scanonly"`
	Status               Status    `bun:"status"`
	CreatedAt            time.Time `bun:"created_at,notnull"`
	UpdatedAt            time.Time `bun:"updated_at,nullzero"`
}

func newAccountModel(acc Account) accountModel {
	return accountModel{
		ID:       acc.ID,
		Name:     acc.Name,
		Agency:   acc.Agency,
		Number:   acc.Number,
		HolderID: acc.HolderID,
		Status:   acc.Status,
	}
}

type accountFilter struct {
	ID             uuid.NullUUID
	Name           string
	DocumentNumber string
	CreatedAtBegin database.NullTime
	CreatedAtEnd   database.NullTime
	UpdatedAtBegin database.NullTime
	UpdatedAtEnd   database.NullTime
}
