package accountsh

import (
	"net/http"

	"github.com/dalmarcogd/ledger-exp/internal/accounts"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type (
	CreateAccountFunc echo.HandlerFunc

	createAccount struct {
		Name           string `json:"name"`
		DocumentNumber string `json:"document_number"`
	}
	createdAccount struct {
		ID             string `json:"id"`
		Name           string `json:"name"`
		Agency         string `json:"agency"`
		Number         string `json:"number"`
		DocumentNumber string `json:"document_number"`
		Status         string `json:"status"`
	}
)

func (c createAccount) Validate() error {
	return validation.ValidateStruct(&c,
		validation.Field(&c.Name, validation.Required, validation.Length(1, 100)),
		validation.Field(&c.DocumentNumber, validation.Required, validation.Length(11, 14)),
	)
}

func NewCreateAccountFunc(svc accounts.Service) CreateAccountFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var acc createAccount
		if err := c.Bind(&acc); err != nil {
			zapctx.L(ctx).Error("create_account_handler_bind_error", zap.Error(err))
			return err
		}

		if err := acc.Validate(); err != nil {
			zapctx.L(ctx).Error("create_holder_handler_validation_error", zap.Error(err))
			return err
		}

		account, err := svc.Create(ctx, accounts.Account{
			Name:           acc.Name,
			DocumentNumber: acc.DocumentNumber,
		})
		if err != nil {
			zapctx.L(ctx).Error("create_account_handler_service_error", zap.Error(err))
			return err
		}

		return c.JSON(
			http.StatusCreated,
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
