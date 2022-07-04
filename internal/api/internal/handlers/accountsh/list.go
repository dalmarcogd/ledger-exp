package accountsh

import (
	"net/http"

	"github.com/dalmarcogd/ledger-exp/internal/accounts"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type (
	ListAccountsFunc echo.HandlerFunc

	listAccounts struct {
		DocumentNumer string `query:"document_number"`
		Sort          int    `query:"sort"`
		Page          int    `query:"page"`
		Size          int    `query:"size"`
	}

	pagination struct {
		Sort        int `json:"sort"`
		Page        int `json:"page"`
		Size        int `json:"size"`
		TotalItems  int `json:"total_items"`
		TotalPages  int `json:"total_pages"`
		TotalInPage int `json:"total_in_page"`
	}

	listedHolder struct {
		Pagination pagination       `json:"pagination"`
		Accounts   []createdAccount `json:"accounts"`
	}
)

func NewListAccountsFunc(svc accounts.Service) ListAccountsFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var lsa listAccounts
		if err := c.Bind(&lsa); err != nil {
			zapctx.L(ctx).Error("list_account_handler_bind_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}

		if lsa.Page == 0 {
			lsa.Page = 1
		}

		if lsa.Size == 0 {
			lsa.Size = 20
		}

		total, hdlrs, err := svc.List(ctx, accounts.ListFilter{
			Sort:           lsa.Sort,
			Page:           lsa.Page,
			Size:           lsa.Size,
			DocumentNumber: lsa.DocumentNumer,
		})
		if err != nil {
			zapctx.L(ctx).Error("list_account_handler_service_error", zap.Error(err))
			return err
		}

		totalPages := total / lsa.Size
		if (total % lsa.Size) != 0 {
			totalPages++
		}

		caccounts := make([]createdAccount, len(hdlrs))
		for i, account := range hdlrs {
			caccounts[i] = createdAccount{
				ID:             account.ID.String(),
				Name:           account.Name,
				Agency:         account.Agency,
				Number:         account.Number,
				DocumentNumber: account.DocumentNumber,
				Status:         string(account.Status),
			}
		}

		listed := listedHolder{
			Pagination: pagination{
				Sort:        lsa.Sort,
				Page:        lsa.Page,
				Size:        lsa.Size,
				TotalItems:  total,
				TotalPages:  totalPages,
				TotalInPage: len(caccounts),
			},
			Accounts: caccounts,
		}

		return c.JSON(http.StatusOK, listed)
	}
}
