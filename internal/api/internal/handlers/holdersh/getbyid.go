package holdersh

import (
	"net/http"

	"github.com/dalmarcogd/ledger-exp/internal/holders"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type (
	GetByIDHolderFunc echo.HandlerFunc

	getByID struct {
		ID string `param:"id"`
	}
)

func NewGetByIDHolderFunc(svc holders.Service) GetByIDHolderFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var get getByID
		if err := c.Bind(&get); err != nil {
			zapctx.L(ctx).Error("get_by_id_holder_handler_bind_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}

		id, err := uuid.Parse(get.ID)
		if err != nil {
			zapctx.L(ctx).Error("get_by_id_holder_handler_bind_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnprocessableEntity, "invalid id")
		}

		holder, err := svc.GetByID(ctx, id)
		if err != nil {
			zapctx.L(ctx).Error("get_by_id_holder_handler_service_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusBadRequest, err.Error())
		}

		return c.JSON(
			http.StatusOK,
			createdHolder{
				ID:             holder.ID.String(),
				Name:           holder.Name,
				DocumentNumber: holder.DocumentNumber,
			},
		)
	}
}
