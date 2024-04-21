package tax

import (
	"encoding/csv"
	"errors"
	"net/http"
	"strconv"

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

func (h *Handler) CalculateTaxFromTaxFile(c echo.Context) error {
	// Source
	taxFile, err := c.FormFile("taxFile")
	if err != nil {
		return c.JSON(http.StatusBadRequest, &Err{
			Message: "unable to read csv file" + err.Error(),
		})
	}
	src, err := taxFile.Open()
	if err != nil {
		return c.JSON(http.StatusBadRequest, &Err{
			Message: "unable to read csv file",
		})
	}
	defer src.Close()

	reader := csv.NewReader(src)

	records, err := reader.ReadAll()
	if err != nil {
		return c.JSON(http.StatusBadRequest, &Err{
			Message: "unable to read csv file",
		})
	}
	headers := records[0]

	taxDetails := []CalculateTaxDetails{}

	for _, row := range records[1:] {
		taxInfo := TaxInformation{
			Allowances: []Allowance{
				{
					AllowanceType: "donation",
					Amount:        0,
				},
			},
		}
		for ic, col := range row {
			data, err := strconv.ParseFloat(col, 64)
			if err != nil {
				return c.JSON(http.StatusBadRequest, &Err{
					Message: "invalid data type in the csv file",
				})
			}

			if headers[ic] == "totalIncome" {
				taxInfo.TotalIncome = data
			} else if headers[ic] == "wht" {
				taxInfo.WHT = data
			} else if headers[ic] == "allowances" {
				taxInfo.Allowances[0].Amount = data
			}
		}

		taxDetails = append(taxDetails, h.taxCalculator.calculate(taxInfo))
	}

	taxes := []TaxCalculationResponse{}

	for _, td := range taxDetails {
		taxes = append(taxes, TaxCalculationResponse{
			Tax:       td.tax,
			TaxRefund: td.taxRefund,
			TaxLevel:  td.taxLevel,
		})
	}

	return c.JSON(http.StatusOK, struct {
		Taxes []TaxCalculationResponse `json:"taxes"`
	}{
		Taxes: taxes,
	})
}
