package holdersh

import (
	"net/http"

	"github.com/dalmarcogd/ledger-exp/internal/holders"
	"github.com/dalmarcogd/ledger-exp/pkg/zapctx"
	"github.com/labstack/echo/v4"
	"go.uber.org/zap"
)

type (
	ListHoldersFunc echo.HandlerFunc

	listHolders struct {
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
		Pagination pagination      `json:"pagination"`
		Holders    []createdHolder `json:"holders"`
	}
)

func NewListHoldersFunc(svc holders.Service) ListHoldersFunc {
	return func(c echo.Context) error {
		ctx := c.Request().Context()

		var lsa listHolders
		if err := c.Bind(&lsa); err != nil {
			zapctx.L(ctx).Error("list_holder_handler_bind_error", zap.Error(err))
			return echo.NewHTTPError(http.StatusUnprocessableEntity, err.Error())
		}

		if lsa.Page == 0 {
			lsa.Page = 1
		}

		if lsa.Size == 0 {
			lsa.Size = 20
		}

		total, hdlrs, err := svc.List(ctx, holders.ListFilter{
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

		cholders := make([]createdHolder, len(hdlrs))
		for i, holder := range hdlrs {
			cholders[i] = createdHolder{
				ID:             holder.ID.String(),
				Name:           holder.Name,
				DocumentNumber: holder.DocumentNumber,
			}
		}

		listed := listedHolder{
			Pagination: pagination{
				Sort:        lsa.Sort,
				Page:        lsa.Page,
				Size:        lsa.Size,
				TotalItems:  total,
				TotalPages:  totalPages,
				TotalInPage: len(cholders),
			},
			Holders: cholders,
		}

		return c.JSON(http.StatusOK, listed)
	}
}
