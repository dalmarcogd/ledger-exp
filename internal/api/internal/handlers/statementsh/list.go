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
		AccountID      string `param:"id"`
		Sort           int    `query:"sort"`
		Page           int    `query:"page"`
		Size           int    `query:"size"`
		CreatedAtBegin string `query:"created_at_begin"`
		CreatedAtEnd   string `query:"created_at_end"`
	}

	account struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}

	statement struct {
		FromAccount *account  `json:"from_account,omitempty"`
		ToAccount   *account  `json:"to_account,omitempty"`
		Type        string    `json:"type"`
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

		var createdAtBegin, createdAtEnd time.Time
		if lsa.CreatedAtBegin != "" {
			createdAtBegin, err = time.Parse("2006-01-02", lsa.CreatedAtBegin)
			if err != nil {
				zapctx.L(ctx).Error("list_account_handler_bind_error", zap.Error(err))
				return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid created_at_begin")
			}
		}

		if lsa.CreatedAtEnd != "" {
			createdAtEnd, err = time.Parse("2006-01-02", lsa.CreatedAtEnd)
			if err != nil {
				zapctx.L(ctx).Error("list_account_handler_bind_error", zap.Error(err))
				return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid created_at_end")
			}
		}

		total, stats, err := svc.List(ctx, statements.ListFilter{
			Sort:           lsa.Sort,
			Page:           lsa.Page,
			Size:           lsa.Size,
			AccountID:      id,
			CreatedAtBegin: createdAtBegin,
			CreatedAtEnd:   createdAtEnd,
		})
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
				Type:      transaction.Type,
				Amount:    transaction.Amount,
				CreatedAt: transaction.CreatedAt,
			}
			if transaction.FromAccount.ID != uuid.Nil {
				accountStatements[i].FromAccount = &account{
					ID:   transaction.FromAccount.ID.String(),
					Name: transaction.FromAccount.Name,
				}
			}
			if transaction.ToAccount.ID != uuid.Nil {
				accountStatements[i].ToAccount = &account{
					ID:   transaction.ToAccount.ID.String(),
					Name: transaction.ToAccount.Name,
				}
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
