package tax

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

type (
	taxCalculationRequest struct {
		TotalIncome float64 `json:"totalIncome"`
		WHT         float64 `json:"wht"`
		Allowances  []struct {
			AllowanceType string  `json:"allowanceType"`
			Amount        float64 `json:"amount"`
		} `json:"allowances"`
	}

	taxCalculationResponse struct {
		Tax float64 `json:"tax"`
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

func (h *Handler) CalculateTax(c echo.Context) error {
	// in the handler for /users?id=<userID>
	var req taxCalculationRequest
	err := c.Bind(&req)
	if err != nil {
		return c.JSON(http.StatusBadRequest, &Err{
			Message: "invalid request body",
		})
	}

	tax := h.taxCalculator.calculate(req.TotalIncome, req.WHT)

	return c.JSON(http.StatusOK, taxCalculationResponse{Tax: tax})
}
