package statements

import (
	"time"

	"github.com/dalmarcogd/ledger-exp/internal/accounts"
)

type Statement struct {
	FromAccount accounts.Account
	ToAccount   accounts.Account
	Amount      float64
	Description string
	CreatedAt   time.Time
}
