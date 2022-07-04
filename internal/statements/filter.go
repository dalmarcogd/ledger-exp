package statements

import (
	"time"

	"github.com/google/uuid"
)

type ListFilter struct {
	Sort           int
	Page           int
	Size           int
	AccountID      uuid.UUID
	CreatedAtBegin time.Time
	CreatedAtEnd   time.Time
}
