package transactions

import (
	"context"

	"github.com/google/uuid"
)

type Service interface {
	Create(ctx context.Context, transaction Transaction) (Transaction, error)
	GetByID(ctx context.Context, id uuid.UUID) (Transaction, error)
}
