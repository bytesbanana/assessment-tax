package tax

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
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

func TestTotalIncomeTaxCalculation(t *testing.T) {
	testCases := []TestCase{
		{
			income:      210_000,
			expectedTax: 0,
		},
		{
			income:      210_001,
			expectedTax: 0.1,
		},
		{
			income:      500_000,
			expectedTax: 29_000,
		},
		{
			income:      560_000,
			expectedTax: 35_000,
		},
		{
			income:      560_001,
			expectedTax: 35_000.15,
		},
		{
			income:      1_060_000,
			expectedTax: 110_000,
		},
		{
			income:      1_060_001,
			expectedTax: 110_000.2,
		},
		{
			income:      2_060_000,
			expectedTax: 310_000,
		},
		{
			income:      2_060_001,
			expectedTax: 310_000.35,
		},
		{
			income:      4_000_000,
			expectedTax: 989_000,
		},
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

				res := &taxCalculationResponse{}
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
	testCases := []TestCase{
		{
			income:      500_000,
			wht:         25_000,
			expectedTax: 4000,
		}, {
			income:      560_000,
			wht:         10_000,
			expectedTax: 25_000,
		}, {
			income:      560_001,
			wht:         10_000,
			expectedTax: 25_000.15,
		}, {
			income:      560_001,
			wht:         10_000,
			expectedTax: 25_000.15,
		}, {
			income:      1_060_000,
			wht:         10_000,
			expectedTax: 100_000,
		},
		{
			income:      1_060_001,
			wht:         10_000,
			expectedTax: 100_000.2,
		},
		{
			income:      2_060_000,
			wht:         10_000,
			expectedTax: 300_000,
		},
		{
			income:      2_060_001,
			wht:         10_000,
			expectedTax: 300_000.35,
		},
		{
			income:      4_000_000,
			wht:         10_000,
			expectedTax: 979_000,
		},
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

				res := &taxCalculationResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				if res.Tax != tc.expectedTax {
					t.Errorf("invalid tax: got %v want %v",
						res.Tax, tc.expectedTax)
				}
			})

		}
	})

}

func sumAllowances(allowances []Allowance) float64 {
	sum := 0.0
	for _, allowance := range allowances {
		sum += allowance.Amount
	}
	return sum
}

func TestTotalIncomeWithAllowancesTaxCalculation(t *testing.T) {
	testCases := []TestCase{
		{
			income: 500_000,
			wht:    0,
			allowances: []Allowance{
				{
					AllowanceType: "donation",
					Amount:        100_000,
				},
			},
			expectedTax: 19_000,
		},
		{
			income: 500_000,
			wht:    0,
			allowances: []Allowance{
				{
					AllowanceType: "donation",
					Amount:        50_000,
				}, {
					AllowanceType: "donation",
					Amount:        50_000,
				},
			},
			expectedTax: 19_000,
		},
		{
			income: 250_000,
			wht:    0,
			allowances: []Allowance{
				{
					AllowanceType: "donation",
					Amount:        20_000,
				}, {
					AllowanceType: "k-receipt",
					Amount:        20_000,
				},
			},
			expectedTax: 0,
		},
		{
			income: 250_001,
			wht:    0,
			allowances: []Allowance{
				{
					AllowanceType: "donation",
					Amount:        20_000,
				}, {
					AllowanceType: "k-receipt",
					Amount:        20_000,
				},
			},
			expectedTax: 0.1,
		}, {
			income: 600_000,
			wht:    0,
			allowances: []Allowance{
				{
					AllowanceType: "donation",
					Amount:        20_000,
				}, {
					AllowanceType: "k-receipt",
					Amount:        20_000,
				},
			},
			expectedTax: 35000,
		},
		{
			income: 600_001,
			wht:    0,
			allowances: []Allowance{
				{
					AllowanceType: "donation",
					Amount:        20_000,
				}, {
					AllowanceType: "k-receipt",
					Amount:        20_000,
				},
			},
			expectedTax: 35000.15,
		},
		{
			income: 1_100_000,
			wht:    0,
			allowances: []Allowance{
				{
					AllowanceType: "donation",
					Amount:        20_000,
				}, {
					AllowanceType: "k-receipt",
					Amount:        20_000,
				},
			},
			expectedTax: 110_000,
		},
		{
			income: 1_100_001,
			wht:    0,
			allowances: []Allowance{
				{
					AllowanceType: "donation",
					Amount:        20_000,
				}, {
					AllowanceType: "k-receipt",
					Amount:        20_000,
				},
			},
			expectedTax: 110_000.2,
		},
		{
			income: 2_100_000,
			wht:    0,
			allowances: []Allowance{
				{
					AllowanceType: "donation",
					Amount:        20_000,
				}, {
					AllowanceType: "k-receipt",
					Amount:        20_000,
				},
			},
			expectedTax: 310000,
		},
		{
			income: 2_100_001,
			wht:    0,
			allowances: []Allowance{
				{
					AllowanceType: "donation",
					Amount:        20_000,
				}, {
					AllowanceType: "k-receipt",
					Amount:        20_000,
				},
			},
			expectedTax: 310000.35,
		},
		{
			income: 4_000_000,
			wht:    0,
			allowances: []Allowance{
				{
					AllowanceType: "donation",
					Amount:        20_000,
				}, {
					AllowanceType: "k-receipt",
					Amount:        20_000,
				},
			},
			expectedTax: 975000,
		},
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

				res := &taxCalculationResponse{}
				json.Unmarshal(rec.Body.Bytes(), res)

				if res.Tax != tc.expectedTax {
					t.Errorf("invalid tax: got %v want %v",
						res.Tax, tc.expectedTax)
				}
			})
		}
	})

}
