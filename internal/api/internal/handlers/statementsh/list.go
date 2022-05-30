package statementsh

import (
	"net/http"
	"time"

	"github.com/dalmarcogd/ledger-exp/internal/statements"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type (
	ListAccountStatementFunc echo.HandlerFunc

	listAccountStatement struct {
		AccountID string `param:"id"`
		Sort      int    `query:"sort"`
		Page      int    `query:"page"`
		Size      int    `query:"size"`
	}

	account struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	statement struct {
		FromAccount account   `json:"from_account"`
		ToAccount   account   `json:"to_account"`
		Amount      float64   `json:"amount"`
		CreatedAt   time.Time `json:"created_at"`
	}

	pagination struct {
		Sort        int `json:"sort"`
		Page        int `json:"page"`
		Size        int `json:"size"`
		TotalItems  int `json:"total_items"`
		TotalPages  int `json:"total_pages"`
		TotalInPage int `json:"total_in_page"`
	}

	listedAccountStatement struct {
		Pagination pagination  `json:"pagination"`
		AccountID  uuid.UUID   `json:"account_id"`
		Statements []statement `json:"statements"`
	}
)

func NewListAccountStatementFunc(svc statements.Service) ListAccountStatementFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var lsa listAccountStatement
		if err := c.Bind(&lsa); err != nil {
			zapctx.L(ctx).Error("list_account_handler_bind_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}

		id, err := uuid.Parse(lsa.AccountID)
		if err != nil {
			zapctx.L(ctx).Error("list_account_handler_bind_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid id")
		}

		if lsa.Page == 0 {
			lsa.Page = 1
		}

		if lsa.Size == 0 {
			lsa.Size = 20
		}

		total, stats, err := svc.List(ctx, id, lsa.Page, lsa.Size, lsa.Sort)
		if err != nil {
			zapctx.L(ctx).Error("list_account_handler_service_error", zap.Error(err))
			return err
		}

		totalPages := total / lsa.Size
		if (total % lsa.Size) != 0 {
			totalPages++
		}

		accountStatements := make([]statement, len(stats))
		for i, transaction := range stats {
			accountStatements[i] = statement{
				FromAccount: account{
					ID:   transaction.FromAccount.ID.String(),
					Name: transaction.FromAccount.Name,
				},
				ToAccount: account{
					ID:   transaction.ToAccount.ID.String(),
					Name: transaction.ToAccount.Name,
				},
				Amount:    transaction.Amount,
				CreatedAt: transaction.CreatedAt,
			}
		}

		listed := listedAccountStatement{
			Pagination: pagination{
				Sort:        lsa.Sort,
				Page:        lsa.Page,
				Size:        lsa.Size,
				TotalItems:  total,
				TotalPages:  totalPages,
				TotalInPage: len(accountStatements),
			},
			AccountID:  id,
			Statements: accountStatements,
		}

		return c.JSON(http.StatusOK, listed)
	}
}
