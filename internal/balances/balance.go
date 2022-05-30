package balances

import (
	"time"

	"github.com/google/uuid"
)

type AccountBalance struct {
	AccountID      uuid.UUID
	CurrentBalance float64
	LastChangeAt   time.Time
}
