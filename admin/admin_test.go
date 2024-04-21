package admin

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/bytesbanana/assessment-tax/postgres"
	"github.com/labstack/echo/v4"
)

type StubAdminHandler struct {
	Configs map[string]*postgres.TaxConfig
}

func (h *StubAdminHandler) SetTaxConfig(key string, value float64) (*postgres.TaxConfig, error) {
	if h.Configs[key] != nil {
		h.Configs[key].Value = value
		return h.Configs[key], nil
	}

	return nil, errors.New("config not found")
}

func TestSetTaxConfig(t *testing.T) {

	t.Run("given invalid set personal deductionrequest should return 400", func(t *testing.T) {
		e := echo.New()
		req := httptest.NewRequest(http.MethodPost, "/admin/deductions/personal-deduction", strings.NewReader("{}"))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		handler := New(&StubAdminHandler{
			Configs: map[string]*postgres.TaxConfig{
				"PERSONAL_DEDUCTION": {
					Key:   "PERSONAL_DEDUCTION",
					Name:  "Personal Deduction",
					Value: 60_000,
				},
			},
		})

		handler.SetPersonalDeductionsConfig(c)

		if rec.Code != http.StatusBadRequest {
			t.Errorf("invalid http status: got %v want %v",
				rec.Code, http.StatusBadRequest)
		}

	})

	t.Run("given new personal deduction amount should update personal deduction", func(t *testing.T) {

		e := echo.New()
		reqJSON := fmt.Sprintf(`{"amount": %f}`, 70000.0)
		req := httptest.NewRequest(http.MethodPost, "/admin/deductions/personal-deduction", strings.NewReader(reqJSON))
		req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		stubAdminHandler := &StubAdminHandler{
			Configs: map[string]*postgres.TaxConfig{
				"PERSONAL_DEDUCTION": {
					Key:   "PERSONAL_DEDUCTION",
					Name:  "Personal Deduction",
					Value: 60_000,
				},
			},
		}
		handler := New(stubAdminHandler)

		handler.SetPersonalDeductionsConfig(c)

		if rec.Code != http.StatusOK {
			t.Errorf("invalid http status: got %v want %v",
				rec.Code, http.StatusOK)
		}

		if stubAdminHandler.Configs["PERSONAL_DEDUCTION"].Value != 70000.0 {
			t.Errorf("invalid personal deduction amount: got %v want %v",
				stubAdminHandler.Configs["PERSONAL_DEDUCTION"].Value, 70000.0)
		}

	})
}