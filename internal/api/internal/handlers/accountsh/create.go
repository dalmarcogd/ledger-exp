package accountsh

import (
	"net/http"

	"github.com/dalmarcogd/ledger-exp/internal/accounts"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type (
	CreateAccountFunc echo.HandlerFunc

	createAccount struct {
		Name string `json:"name"`
	}
	createdAccount struct {
		ID   string `json:"id"`
		Name string `json:"name"`
	}
)

func NewCreateAccountFunc(svc accounts.Service) CreateAccountFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var acc createAccount
		if err := c.Bind(&acc); err != nil {
			zapctx.L(ctx).Error("create_account_handler_bind_error", zap.Error(err))
			return err
		}

		account, err := svc.Create(ctx, accounts.Account{
			Name: acc.Name,
		})
		if err != nil {
			zapctx.L(ctx).Error("create_account_handler_service_error", zap.Error(err))
			return err
		}

		return c.JSON(
			http.StatusCreated,
			createdAccount{
				ID:   account.ID.String(),
				Name: account.Name,
			},
		)
	}
}
