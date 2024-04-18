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
	t.Run("given total income 210,001 should return tax 0.1", func(t *testing.T) {
		c, rec := setup(t, func() *http.Request {
			reqJSON := `{
				"totalIncome": 210001.0,
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

		if res.Tax != 0.1 {
			t.Errorf("invalid tax: got %v want %v",
				res.Tax, 0.1)
		}
	})

	t.Run("given total income 560,000 should return tax 35,000", func(t *testing.T) {
		c, rec := setup(t, func() *http.Request {
			reqJSON := `{
				"totalIncome": 560000.0,
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

		if res.Tax != 35000 {
			t.Errorf("invalid tax: got %v want %v",
				res.Tax, 35000)
		}
	})

	t.Run("given total income 560,001 should return tax 35,000.15", func(t *testing.T) {
		c, rec := setup(t, func() *http.Request {
			reqJSON := `{
				"totalIncome": 560001.0,
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

		if res.Tax != 35000.15 {
			t.Errorf("invalid tax: got %v want %v",
				res.Tax, 35000.15)
		}
	})

	t.Run("given total income 1,060,000 should return tax 110,000", func(t *testing.T) {
		c, rec := setup(t, func() *http.Request {
			reqJSON := `{
				"totalIncome": 1060000.0,
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

		if res.Tax != 110000 {
			t.Errorf("invalid tax: got %v want %v",
				res.Tax, 110000)
		}
	})
}
