package statements

import (
	"context"

	"github.com/dalmarcogd/ledger-exp/pkg/database"
	"github.com/dalmarcogd/ledger-exp/pkg/tracer"
)

type Repository interface {
	ListByFilter(ctx context.Context, filter StatementFilter) (int, []statementModel, error)
}

type repository struct {
	tracer tracer.Tracer
	db     database.Database
}

func NewRepository(t tracer.Tracer, db database.Database) Repository {
	return repository{
		tracer: t,
		db:     db,
	}
}

func (r repository) ListByFilter(ctx context.Context, filter StatementFilter) (int, []statementModel, error) {
	ctx, span := r.tracer.Span(ctx)
	defer span.End()

	page := filter.Page
	if page == 0 {
		page = 1
	}

	size := filter.Size
	if size == 0 {
		size = 20
	}

	selectQuery := r.db.Replica().
		NewSelect().
		Model(&statementModel{}).
		ColumnExpr("trx.*").
		ColumnExpr("from_acc.name AS from_account_name, to_acc.name AS to_account_name").
		Where(
			"(from_account_id = ? OR to_account_id = ?)",
			filter.AccountID.String(),
			filter.AccountID.String(),
		).
		Join("LEFT JOIN accounts AS from_acc").
		JoinOn("from_acc.id = from_account_id").
		Join("LEFT JOIN accounts AS to_acc").
		JoinOn("to_acc.id = to_account_id").
		Limit(size).
		Offset((page - 1) * size)

	if filter.Sort == 0 {
		selectQuery.Order("trx.created_at ASC")
	} else if filter.Sort > 0 {
		selectQuery.Order("trx.created_at DESC")
	}

	if !filter.CreatedAtBegin.IsZero() {
		selectQuery.Where("trx.created_at >= ?", filter.CreatedAtBegin)
	}
	if !filter.CreatedAtEnd.IsZero() {
		selectQuery.Where("trx.created_at <= ?", filter.CreatedAtEnd)
	}

	var stms []statementModel
	total, err := selectQuery.ScanAndCount(ctx, &stms)
	if err != nil {
		span.RecordError(err)
		return 0, []statementModel{}, err
	}

	return total, stms, nil
}
