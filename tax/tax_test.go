package tax

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/labstack/echo/v4"
)

type TestCase struct {
	income      float64
	wht         float64
	allowances  []Allowance
	expectedTax float64
}

func setup(t *testing.T, buildRequestFunc func() *http.Request) (echo.Context, *httptest.ResponseRecorder) {
	t.Parallel()
	e := echo.New()
	req := buildRequestFunc()
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/wallets")

	return c, rec
}

func TestRequestValidtion(t *testing.T) {

	t.Run("given invalid request should return 400", func(t *testing.T) {
		c, rec := setup(t, func() *http.Request {
			reqJSON := `{
				"salary": 500000.0,
			}`
			return httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqJSON))
		})

		h := &Handler{}
		h.CalculateTax(c)

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
		h.CalculateTax(c)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("invalid http status: got %v want %v",
				rec.Code, http.StatusBadRequest)
		}
	})
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

func TestTotalIncomeTaxCalculation(t *testing.T) {
	testCases, err := loadTestCasesFromFile("./data/income_test_data.json")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("total income calculation", func(t *testing.T) {

		for _, tc := range testCases {
			name := fmt.Sprintf("given total income %.2f should return tax amount %.2f", tc.income, tc.expectedTax)
			t.Run(name, func(t *testing.T) {

				c, rec := setup(t, func() *http.Request {
					reqJSON := fmt.Sprintf(`{"totalIncome": %f}`, tc.income)
					return httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqJSON))
				})

				h := &Handler{}
				h.CalculateTax(c)

				if rec.Code != http.StatusOK {
					t.Errorf("invalid status code: got %v want %v",
						rec.Code, http.StatusOK)
				}

				res := &TaxCalculationResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				if res.Tax != tc.expectedTax {
					t.Errorf("invalid tax: got %v want %v",
						res.Tax, tc.expectedTax)
				}
			})

		}
	})

}

func TestTotalIncomeWHTTaxCalculation(t *testing.T) {
	testCases, err := loadTestCasesFromFile("./data/income_wth_test_data.json")
	if err != nil {
		t.Fatal(err)
	}

	t.Run("total income + WHT calculation", func(t *testing.T) {

		for _, tc := range testCases {
			name := fmt.Sprintf("given total income %.2f and WHT %.2f should return tax amount %.2f",
				tc.income,
				tc.wht,
				tc.expectedTax)
			t.Run(name, func(t *testing.T) {

				c, rec := setup(t, func() *http.Request {
					reqJSON := fmt.Sprintf(`{
						"totalIncome": %f,
						"wht": %f
					}`, tc.income, tc.wht)
					return httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqJSON))
				})

				h := &Handler{}
				h.CalculateTax(c)

				if rec.Code != http.StatusOK {
					t.Errorf("invalid status code: got %v want %v",
						rec.Code, http.StatusOK)
				}

				res := &TaxCalculationResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				if res.Tax != tc.expectedTax {
					t.Errorf("invalid tax: got %v want %v",
						res.Tax, tc.expectedTax)
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

			allowances, err := json.Marshal(tc.allowances)
			if err != nil {
				t.Errorf("invalid allowances: %v", err)
				return
			}

			name := fmt.Sprintf("given total income %.2f and allowances %.2f should return tax amount %.2f",
				tc.income,
				sumAllowances(tc.allowances),
				tc.expectedTax)

			t.Run(name, func(t *testing.T) {

				c, rec := setup(t, func() *http.Request {
					reqJSON := fmt.Sprintf(`{
						"totalIncome": %f,
						"wht": %f,
						"allowances": %s
					}`,
						tc.income,
						tc.wht,
						allowances,
					)
					return httptest.NewRequest(http.MethodPost, "/", strings.NewReader(reqJSON))
				})

				h := &Handler{}
				h.CalculateTax(c)

				if rec.Code != http.StatusOK {
					t.Errorf("invalid status code: got %v want %v",
						rec.Code, http.StatusOK)
				}

				res := &TaxCalculationResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				if res.Tax != tc.expectedTax {
					t.Errorf("invalid tax: got %v want %v",
						res.Tax, tc.expectedTax)
				}
			})
		}
	})

}
