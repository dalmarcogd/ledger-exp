package transactionsh

import (
	"net/http"

	"github.com/dalmarcogd/ledger-exp/internal/transactions"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type (
	GetByIDTransactionFunc echo.HandlerFunc

	getByID struct {
		ID string `param:"id"`
	}
)

func NewGetByIDTransactionFunc(svc transactions.Service) GetByIDTransactionFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var get getByID
		if err := c.Bind(&get); err != nil {
			zapctx.L(ctx).Error("get_by_id_transactions_handler_bind_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}

		id, err := uuid.Parse(get.ID)
		if err != nil {
			zapctx.L(ctx).Error("get_by_id_transactions_handler_bind_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid id")
		}

		transaction, err := svc.GetByID(ctx, id)
		if err != nil {
			zapctx.L(ctx).Error("get_by_id_transactions_handler_service_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return c.JSON(
			http.StatusOK,
			createdTransaction{
				ID:          transaction.ID.String(),
				From:        transaction.From.String(),
				To:          transaction.To.String(),
				Amount:      float64(transaction.Amount),
				Description: transaction.Description,
			},
		)
	}
}
