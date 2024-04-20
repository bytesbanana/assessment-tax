package tax

import (
	"errors"
	"net/http"

	"github.com/labstack/echo/v4"
)

var ACCEPT_ALLOWANCE_TYPES = map[string]string{
	"k-receipt":  "k-receipt",
	"donation":   "donation",
	"e-shopping": "e-shopping",
}

type (
	Allowance struct {
		AllowanceType string  `json:"allowanceType"`
		Amount        float64 `json:"amount"`
	}

	TaxInformation struct {
		TotalIncome float64     `json:"totalIncome"`
		WHT         float64     `json:"wht"`
		Allowances  []Allowance `json:"allowances"`
	}

	TaxLevel struct {
		Level string  `json:"level"`
		Tax   float64 `json:"tax"`
	}

	TaxCalculationResponse struct {
		Tax      float64    `json:"tax"`
		TaxLevel []TaxLevel `json:"taxLevel"`
	}

	Handler struct {
		taxCalculator TaxCalculator
	}

	Err struct {
		Message string `json:"message"`
	}
)

func NewHandler() *Handler {
	return &Handler{
		taxCalculator: New(),
	}
}

func validateAllowance(allowances []Allowance) error {
	for _, al := range allowances {
		if ACCEPT_ALLOWANCE_TYPES[al.AllowanceType] == "" {
			return errors.New("invalid allowance type")
		}
	}
	return nil
}

func (h *Handler) CalculateTax(c echo.Context) error {

	var req TaxInformation
	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &Err{
			Message: "invalid request body",
		})
	}

	err = validateAllowance(req.Allowances)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &Err{
			Message: "invalid  allowance type",
		})
	}

	taxDetails := h.taxCalculator.calculate(req)

	return c.JSON(http.StatusOK, TaxCalculationResponse{
		Tax:      taxDetails.tax,
		TaxLevel: taxDetails.taxLevel,
	})
}
