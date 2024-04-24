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
		storer Storer
	}

	Err struct {
		Message string `json:"message"`
	}
)

func New(db Storer) *Handler {

	return &Handler{
		storer: db,
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

	personalDeducation := h.getConfigValue("PERSONAL_DEDUCTION", 60_000)
	maxKReceiptDeduction := h.getConfigValue("MAX_K_RECEIPT_DEDUCTION", 50_000)
	taxCalculator := NewTaxCalculator(personalDeducation, maxKReceiptDeduction)

	taxDetails := taxCalculator.calculate(req)

	return c.JSON(http.StatusOK, TaxCalculationResponse{
		Tax:       taxDetails.tax,
		TaxRefund: taxDetails.taxRefund,
		TaxLevel:  taxDetails.taxLevel,
	})
}

func (h *Handler) getConfigValue(configName string, defaultValue float64) float64 {
	result := defaultValue

	config, err := h.storer.GetTaxConfig(configName)
	if err == nil {
		result = config.Value
	}

	return result
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

	personalDeducation := h.getConfigValue("PERSONAL_DEDUCTION", 60_000)
	maxKReceiptDeduction := h.getConfigValue("MAX_K_RECEIPT_DEDUCTION", 50_000)
	taxCalculator := NewTaxCalculator(personalDeducation, maxKReceiptDeduction)

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

		taxDetails = append(taxDetails, taxCalculator.calculate(taxInfo))
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
