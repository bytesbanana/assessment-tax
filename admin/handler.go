package admin

import (
	"net/http"

	"github.com/bytesbanana/assessment-tax/postgres"
	"github.com/labstack/echo/v4"
)

type (
	Storer interface {
		SetTaxConfig(key string, value float64) (*postgres.TaxConfig, error)
	}

	Handler struct {
		store Storer
	}

	SetPersonalDeductionRequest struct {
		Amount *float64 `json:"amount,omitempty" validate:"required"`
	}

	Err struct {
		Message string `json:"message"`
	}
)

func New(db Storer) *Handler {

	return &Handler{
		store: db,
	}
}

func (h *Handler) SetPersonalDeductionsConfig(c echo.Context) error {
	var req SetPersonalDeductionRequest
	var err error

	err = c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &Err{
			Message: "invalid request body",
		})
	}

	if req.Amount == nil {
		return c.JSON(http.StatusBadRequest, &Err{
			Message: "invalid request body",
		})
	}

	personalDeduction, err := h.store.SetTaxConfig("PERSONAL_DEDUCTION", *req.Amount)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &Err{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, struct {
		PersonalDeduction float64 `json:"personalDeduction"`
	}{
		PersonalDeduction: personalDeduction.Value,
	})
}
