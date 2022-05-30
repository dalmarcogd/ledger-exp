package statements

import (
	"context"

	"github.com/dalmarcogd/ledger-exp/internal/accounts"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

type Service interface {
	List(ctx context.Context, accountID uuid.UUID, page, size, sort int) (int, []Statement, error)
}

type service struct {
	tracer     tracer.Tracer
	repository Repository
}

func NewService(t tracer.Tracer, r Repository) Service {
	return service{tracer: t, repository: r}
}

func (s service) List(ctx context.Context, accountID uuid.UUID, page, size, sort int) (int, []Statement, error) {
	ctx, span := s.tracer.Span(ctx)
	defer span.End()

	total, statementModels, err := s.repository.ListByFilter(ctx, statementFilter{
		Page:      page,
		Size:      size,
		Sort:      sort,
		AccountID: accountID,
	})
	if err != nil {
		zapctx.L(ctx).Error("statements_service_repository_error", zap.Error(err))
		span.RecordError(err)
		return 0, []Statement{}, err
	}

	stmts := make([]Statement, len(statementModels))
	for i, model := range statementModels {
		stmts[i] = Statement{
			FromAccount: accounts.Account{
				ID:   model.FromAccountID,
				Name: model.FromAccountName,
			},
			ToAccount: accounts.Account{
				ID:   model.ToAccountID,
				Name: model.ToAccountName,
			},
			Amount:      model.Amount,
			Description: model.Description,
			CreatedAt:   model.CreatedAt,
		}
	}

	return total, stmts, nil
}
