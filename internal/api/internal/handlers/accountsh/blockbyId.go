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
	BlockByIDFunc echo.HandlerFunc

	blockByID struct {
		ID string `param:"id"`
	}
)

func NewBlockByIDFunc(svc accounts.Service) BlockByIDFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var cls blockByID
		if err := c.Bind(&cls); err != nil {
			zapctx.L(ctx).Error("block_by_account_id_handler_bind_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}

		id, err := uuid.Parse(cls.ID)
		if err != nil {
			zapctx.L(ctx).Error("block_by_account_id_handler_bind_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid id")
		}

		account, err := svc.BlockByID(ctx, id)
		if err != nil {
			zapctx.L(ctx).Error("block_by_account_id_handler_service_error", zap.Error(err))
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
