package balances

import (
	"github.com/google/uuid"
)

type AccountBalance struct {
	AccountID      uuid.UUID
	CurrentBalance float64
}
