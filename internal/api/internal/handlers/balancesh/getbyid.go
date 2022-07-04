package balancesh

import (
	"net/http"

	"github.com/dalmarcogd/ledger-exp/internal/balances"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type (
	GetBalanceByAccountIDFunc echo.HandlerFunc

	getBalanceByAccountID struct {
		ID string `param:"id"`
	}
	accountBalance struct {
		AccountID      string  `json:"account_id"`
		CurrentBalance float64 `json:"current_balance"`
	}
)

func NewGetBalanceByAccountIDFunc(svc balances.Service) GetBalanceByAccountIDFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var get getBalanceByAccountID
		if err := c.Bind(&get); err != nil {
			zapctx.L(ctx).Error("get_balance_by_account_id_handler_bind_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}

		id, err := uuid.Parse(get.ID)
		if err != nil {
			zapctx.L(ctx).Error("get_balance_by_account_id_handler_bind_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid id")
		}

		accb, err := svc.GetByAccountID(ctx, id)
		if err != nil {
			zapctx.L(ctx).Error("get_balance_by_account_id_handler_service_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return c.JSON(
			http.StatusOK,
			accountBalance{
				AccountID:      accb.AccountID.String(),
				CurrentBalance: accb.CurrentBalance,
			},
		)
	}
}
