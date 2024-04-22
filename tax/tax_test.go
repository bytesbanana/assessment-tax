package tax

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/bytesbanana/assessment-tax/postgres"
	"github.com/labstack/echo/v4"
)

type TestCase struct {
	Income           float64     `json:"income"`
	Wht              float64     `json:"wht"`
	Allowances       []Allowance `json:"allowances"`
	TaxRefund        float64     `json:"taxRefund"`
	ExpectedTax      float64     `json:"expectedTax"`
	ExpectedTaxLevel []TaxLevel  `json:"expectedTaxLevel"`
}

type StubTaxHandler struct {
	configs map[string]*postgres.TaxConfig
}

func (t *StubTaxHandler) GetTaxConfig(key string) (*postgres.TaxConfig, error) {
	return t.configs[key], nil
}

func setup(t *testing.T, buildRequestFunc func() *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	t.Parallel()
	e := echo.New()
	req := buildRequestFunc()
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/tax/calculations")

	return c, rec
}

func loadTestCasesFromFile(filePath string) ([]TestCase, error) {
	var testCases []TestCase
	file, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer file.Close()
	err = json.NewDecoder(file).Decode(&testCases)
	if err != nil {
		return nil, err
	}
	return testCases, nil
}

func sumAllowances(allowances []Allowance) float64 {
	sum := 0.0
	for _, allowance := range allowances {
		sum += allowance.Amount
	}
	return sum
}

func TestRequestValidtion(t *testing.T) {

	t.Run("given invalid request should return 400", func(t *testing.T) {
		c, rec := setup(t, func() *http.Request {
			reqJSON := `{
				"salary": 500000.0,
			}`
			return httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqJSON))
		})

		h := New(&StubTaxHandler{
			configs: map[string]*postgres.TaxConfig{
				"PERSONAL_DEDUCTION": {
					Value: 60_000,
				}, "MAX_K_RECEIPT_DEDUCTION": {
					Value: 50_000,
				},
			},
		})
		err := h.CalculateTax(c)
		if err != nil {
			t.Errorf("unable to calculate tax: %v", err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Errorf("invalid http status: got %v want %v",
				rec.Code, http.StatusBadRequest)
		}
	})

	t.Run("given invalid allowance type request should return 400", func(t *testing.T) {
		c, rec := setup(t, func() *http.Request {
			reqJSON := `{
				"totalIncome": 500000.0,
				"wht": 0.0,
				"allowances": [
				  {
					"allowanceType": "investment",
					"amount": 0.0
				  }
				]
			  }`
			return httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqJSON))
		})

		h := &Handler{}
		err := h.CalculateTax(c)
		if err != nil {
			t.Errorf("unable to calculate tax: %v", err)
		}

		if rec.Code != http.StatusBadRequest {
			t.Errorf("invalid http status: got %v want %v",
				rec.Code, http.StatusBadRequest)
		}
	})
}

func TestTotalIncomeTaxCalculation(t *testing.T) {
	testCases, err := loadTestCasesFromFile("./data/income_test_data.json")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("total income calculation", func(t *testing.T) {

		for _, tc := range testCases {
			name := fmt.Sprintf("given total income %.2f should return tax amount %.2f with tax level", tc.Income, tc.ExpectedTax)
			t.Run(name, func(t *testing.T) {

				c, rec := setup(t, func() *http.Request {
					reqJSON := fmt.Sprintf(`{"totalIncome": %f}`, tc.Income)
					return httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqJSON))
				})

				h := New(&StubTaxHandler{
					configs: map[string]*postgres.TaxConfig{
						"PERSONAL_DEDUCTION": {
							Value: 60_000,
						},
						"MAX_K_RECEIPT_DEDUCTION": {
							Value: 50_000,
						},
					},
				})

				err := h.CalculateTax(c)
				if err != nil {
					t.Errorf("unable to calculate tax: %v", err)
				}

				if rec.Code != http.StatusOK {
					t.Errorf("invalid status code: got %v want %v",
						rec.Code, http.StatusOK)
				}

				res := &TaxCalculationResponse{}
				err = json.Unmarshal(rec.Body.Bytes(), res)
				if err != nil {
					t.Errorf("unable to unmarshal response: %v", err)
				}

				if res.Tax != tc.ExpectedTax {
					t.Errorf("invalid tax: got %v want %v",
						res.Tax, tc.ExpectedTax)
				}

				if res.TaxRefund != tc.TaxRefund {
					t.Errorf("invalid tax refund: got %v want %v",
						res.TaxRefund, tc.TaxRefund)
				}

				if !reflect.DeepEqual(res.TaxLevel, tc.ExpectedTaxLevel) {
					t.Errorf("invalid tax level: got %v want %v",
						res.Tax, tc.ExpectedTax)
				}
			})

		}
	})

}

func TestTotalIncomeWHTTaxCalculation(t *testing.T) {
	testCases, err := loadTestCasesFromFile("./data/income_wht_test_data.json")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("total income + WHT calculation", func(t *testing.T) {

		for _, tc := range testCases {
			name := fmt.Sprintf("given total income %.2f and WHT %.2f should return tax amount %.2f",
				tc.Income,
				tc.Wht,
				tc.ExpectedTax)
			t.Run(name, func(t *testing.T) {

				c, rec := setup(t, func() *http.Request {
					reqJSON := fmt.Sprintf(`{
						"totalIncome": %f,
						"wht": %f
					}`, tc.Income, tc.Wht)
					return httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqJSON))
				})

				h := New(&StubTaxHandler{
					configs: map[string]*postgres.TaxConfig{
						"PERSONAL_DEDUCTION": {
							Value: 60_000,
						},
						"MAX_K_RECEIPT_DEDUCTION": {
							Value: 50_000,
						},
					},
				})
				err := h.CalculateTax(c)
				if err != nil {
					t.Errorf("unable to calculate tax: %v", err)
				}

				if rec.Code != http.StatusOK {
					t.Errorf("invalid status code: got %v want %v",
						rec.Code, http.StatusOK)
				}

				res := &TaxCalculationResponse{}
				err = json.Unmarshal(rec.Body.Bytes(), res)
				if err != nil {
					t.Errorf("unable to unmarshal response: %v", err)
				}

				if res.Tax != tc.ExpectedTax {
					t.Errorf("invalid tax: got %v want %v",
						res.Tax, tc.ExpectedTax)
				}

				if res.TaxRefund != tc.TaxRefund {
					t.Errorf("invalid tax refund: got %v want %v",
						res.TaxRefund, tc.TaxRefund)
				}

				if !reflect.DeepEqual(res.TaxLevel, tc.ExpectedTaxLevel) {
					t.Errorf("invalid tax level: got %v want %v",
						res.Tax, tc.ExpectedTax)
				}
			})

		}
	})

}

func TestTotalIncomeWithAllowancesTaxCalculation(t *testing.T) {
	testCases, err := loadTestCasesFromFile("./data/income_allowances_test_data.json")
	if err != nil {
		t.Fatal(err)
	}
	t.Run("total income with allowances", func(t *testing.T) {
		for _, tc := range testCases {

			allowances, err := json.Marshal(tc.Allowances)
			if err != nil {
				t.Errorf("invalid allowances: %v", err)
				return
			}

			name := fmt.Sprintf("given total income %.2f and allowances %.2f should return tax amount %.2f",
				tc.Income,
				sumAllowances(tc.Allowances),
				tc.ExpectedTax)

			t.Run(name, func(t *testing.T) {

				c, rec := setup(t, func() *http.Request {
					reqJSON := fmt.Sprintf(`{
						"totalIncome": %f,
						"wht": %f,
						"allowances": %s
					}`,
						tc.Income,
						tc.Wht,
						allowances,
					)
					return httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqJSON))
				})

				h := New(&StubTaxHandler{
					configs: map[string]*postgres.TaxConfig{
						"PERSONAL_DEDUCTION": {
							Value: 60_000,
						},
						"MAX_K_RECEIPT_DEDUCTION": {
							Value: 50_000,
						},
					},
				})
				err := h.CalculateTax(c)
				if err != nil {
					t.Errorf("unable to calculate tax: %v", err)
				}

				if rec.Code != http.StatusOK {
					t.Errorf("invalid status code: got %v want %v",
						rec.Code, http.StatusOK)
				}

				res := &TaxCalculationResponse{}
				err = json.Unmarshal(rec.Body.Bytes(), res)
				if err != nil {
					t.Errorf("unable to unmarshal response: %v", err)
				}

				if res.Tax != tc.ExpectedTax {
					t.Errorf("invalid tax: got %v want %v",
						res.Tax, tc.ExpectedTax)
				}

				if res.TaxRefund != tc.TaxRefund {
					t.Errorf("invalid tax refund: got %v want %v",
						res.TaxRefund, tc.TaxRefund)
				}

				if !reflect.DeepEqual(res.TaxLevel, tc.ExpectedTaxLevel) {
					t.Errorf("invalid tax level: got %v want %v",
						res.Tax, tc.ExpectedTax)
				}
			})
		}
	})

}
