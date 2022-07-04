package transactionsh

import (
	"errors"
	"net/http"

	"github.com/dalmarcogd/ledger-exp/internal/api/internal/handlers/stringers"
	"github.com/dalmarcogd/ledger-exp/internal/transactions"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type (
	CreateCreditTransactionFunc echo.HandlerFunc

	createCreditTransaction struct {
		To          string  `json:"to_account_id"`
		Amount      float64 `json:"amount"`
		Description string  `json:"description"`
	}
)

func NewCreateCreditTransactionFunc(svc transactions.Service) CreateCreditTransactionFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var trx createCreditTransaction
		err := c.Bind(&trx)
		if err != nil {
			zapctx.L(ctx).Error("create_credit_transaction_handler_bind_error", zap.Error(err))
			return err
		}

		var toID uuid.UUID
		if trx.To != "" {
			toID, err = uuid.Parse(trx.To)
			if err != nil {
				zapctx.L(ctx).Error("create_credit_transaction_handler_parse_error", zap.Error(err))
				return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid to account id")
			}
		}

		transaction, err := svc.CreateCredit(ctx, transactions.Transaction{
			To:          toID,
			Amount:      trx.Amount,
			Description: trx.Description,
		})
		if err != nil {
			zapctx.L(ctx).Error("create_credit_transaction_handler_service_error", zap.Error(err))
			if errors.Is(err, transactions.ErrAccountNotfound) {
				return echo.NewHTTPError(http.StatusBadRequest, err.Error())
			}
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return c.JSON(
			http.StatusCreated,
			createdTransaction{
				ID:          stringers.UUIDEmpty(transaction.ID),
				To:          stringers.UUIDEmpty(transaction.To),
				Type:        string(transaction.Type),
				Amount:      transaction.Amount,
				Description: transaction.Description,
			},
		)
	}
}
