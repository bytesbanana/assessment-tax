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

func TestTaxCalculation(t *testing.T) {

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

	testCases := []struct {
		income      float64
		expectedTax float64
	}{
		{
			income:      210_000,
			expectedTax: 0,
		},
		{
			income:      210_001,
			expectedTax: 0.1,
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
	}

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

}
