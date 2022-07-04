package accountsh

import (
	"net/http"

	"github.com/dalmarcogd/ledger-exp/internal/accounts"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type (
	CloseByIDFunc echo.HandlerFunc

	closeByID struct {
		ID string `param:"id"`
	}
)

func NewCloseByIDFunc(svc accounts.Service) CloseByIDFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var cls closeByID
		if err := c.Bind(&cls); err != nil {
			zapctx.L(ctx).Error("close_by_account_id_handler_bind_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}

		id, err := uuid.Parse(cls.ID)
		if err != nil {
			zapctx.L(ctx).Error("close_by_account_id_handler_bind_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid id")
		}

		account, err := svc.CloseByID(ctx, id)
		if err != nil {
			zapctx.L(ctx).Error("blocse_by_account_id_handler_service_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return c.JSON(
			http.StatusOK,
			createdAccount{
				ID:             account.ID.String(),
				Name:           account.Name,
				Agency:         account.Agency,
				Number:         account.Number,
				DocumentNumber: account.DocumentNumber,
				Status:         string(account.Status),
			},
		)
	}
}
