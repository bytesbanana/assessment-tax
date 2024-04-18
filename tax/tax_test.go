package tax

import (
	"encoding/json"
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

	t.Run("given total income 210,000 should return tax 0", func(t *testing.T) {
		c, rec := setup(t, func() *http.Request {
			reqJSON := `{
				"totalIncome": 210000.0,
				"wht": 0.0,
				"allowances": [
				  {
					"allowanceType": "donation",
					"amount": 0.0
				  }
				]
			}`
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

		if res.Tax != 0.0 {
			t.Errorf("invalid tax: got %v want %v",
				res.Tax, 0.0)
		}

	})
}
