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

	SetConfigValueRequest struct {
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
	var req SetConfigValueRequest
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

	if *req.Amount < 10_000 || *req.Amount > 100_000 {
		return c.JSON(http.StatusBadRequest, &Err{
			Message: "amount must be between 10,000 and 100,000",
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

func (h *Handler) SetMaxKReceiptDeduction(c echo.Context) error {

	var req SetConfigValueRequest
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

	if *req.Amount < 1 || *req.Amount > 100_000 {
		return c.JSON(http.StatusBadRequest, &Err{
			Message: "amount must be between 1 and 100,000",
		})
	}

	maxKReceipt, err := h.store.SetTaxConfig("MAX_K_RECEIPT_DEDUCTION", *req.Amount)
	if err != nil {
		return c.JSON(http.StatusInternalServerError, &Err{
			Message: err.Error(),
		})
	}

	return c.JSON(http.StatusOK, struct {
		KReceipt float64 `json:"kReceipt"`
	}{
		KReceipt: maxKReceipt.Value,
	})
}
