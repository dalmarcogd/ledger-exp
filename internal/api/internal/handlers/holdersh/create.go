package holdersh

import (
	"net/http"

	"github.com/dalmarcogd/ledger-exp/internal/holders"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	"github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type (
	CreateHolderFunc echo.HandlerFunc

	createHolder struct {
		Name           string `json:"name"`
		DocumentNumber string `json:"document_number"`
	}
	createdHolder struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		DocumentNumber string `json:"document_number"`
	}
)

func (c createHolder) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&c.DocumentNumber, validation.Required, validation.Length(11, 14)),
	)
}

func NewCreateHolderFunc(svc holders.Service) CreateHolderFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var acc createHolder
		if err := c.Bind(&acc); err != nil {
			zapctx.L(ctx).Error("create_holder_handler_bind_error", zap.Error(err))
			return err
		}

		if err := acc.Validate(); err != nil {
			zapctx.L(ctx).Error("create_holder_handler_validation_error", zap.Error(err))
			return err
		}

		holder, err := svc.Create(ctx, holders.Holder{
			Name:           acc.Name,
			DocumentNumber: acc.DocumentNumber,
		})
		if err != nil {
			zapctx.L(ctx).Error("create_holder_handler_service_error", zap.Error(err))
			return err
		}

		return c.JSON(
			http.StatusCreated,
			createdHolder{
				ID:             holder.ID.String(),
				Name:           holder.Name,
				DocumentNumber: holder.DocumentNumber,
			},
		)
	}
}
