package tax

import (
	"errors"
	"net/http"

	"github.com/bytesbanana/assessment-tax/postgres"
	"github.com/labstack/echo/v4"
)

var ACCEPT_ALLOWANCE_TYPES = map[string]string{
	"k-receipt": "k-receipt",
	"donation":  "donation",
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
		Tax       float64    `json:"tax"`
		TaxRefund float64    `json:"taxRefund"`
		TaxLevel  []TaxLevel `json:"taxLevel"`
	}

	Storer interface {
		GetTaxConfig(key string) (*postgres.TaxConfig, error)
	}

	Handler struct {
		taxCalculator TaxCalculator
		storer        Storer
	}

	Err struct {
		Message string `json:"message"`
	}
)

func New(db Storer) *Handler {

	personalDededucationConfig, err := db.GetTaxConfig("PERSONAL_DEDUCTION")
	if err != nil {
		return &Handler{
			taxCalculator: NewTaxCalculator(60_000),
			storer:        db,
		}
	}

	return &Handler{
		taxCalculator: NewTaxCalculator(personalDededucationConfig.Value),
		storer:        db,
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
		Tax:       taxDetails.tax,
		TaxRefund: taxDetails.taxRefund,
		TaxLevel:  taxDetails.taxLevel,
	})
}
