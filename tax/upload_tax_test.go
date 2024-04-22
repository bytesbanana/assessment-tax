package tax

import (
	"bytes"
	"encoding/json"
	"io"
	"log"
	"mime/multipart"
	"reflect"

	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/bytesbanana/assessment-tax/postgres"
	"github.com/labstack/echo/v4"
)

func TestTaxFileUpdaloadCalculation(t *testing.T) {
	t.Parallel()

	expectedResult := struct {
		Taxes []TaxCalculationResponse `json:"taxes"`
	}{
		Taxes: []TaxCalculationResponse{
			{
				Tax:       29000,
				TaxRefund: 0,
				TaxLevel: []TaxLevel{
					{
						Level: "0-150,000",
						Tax:   0,
					},
					{
						Level: "150,001-500,000",
						Tax:   29000,
					},
					{
						Level: "500,001-1,000,000",
						Tax:   0,
					},
					{
						Level: "1,000,001-2,000,000",
						Tax:   0,
					},
					{
						Level: "2,000,001 ขึ้นไป",
						Tax:   0,
					},
				},
			},
			{
				Tax:       1000,
				TaxRefund: 0,
				TaxLevel: []TaxLevel{
					{
						Level: "0-150,000",
						Tax:   0,
					},
					{
						Level: "150,001-500,000",
						Tax:   35000,
					},
					{
						Level: "500,001-1,000,000",
						Tax:   6000,
					},
					{
						Level: "1,000,001-2,000,000",
						Tax:   0,
					},
					{
						Level: "2,000,001 ขึ้นไป",
						Tax:   0,
					},
				},
			},
			{
				Tax:       13500,
				TaxRefund: 0,
				TaxLevel: []TaxLevel{
					{
						Level: "0-150,000",
						Tax:   0,
					},
					{
						Level: "150,001-500,000",
						Tax:   35000,
					},
					{
						Level: "500,001-1,000,000",
						Tax:   28500,
					},
					{
						Level: "1,000,001-2,000,000",
						Tax:   0,
					},
					{
						Level: "2,000,001 ขึ้นไป",
						Tax:   0,
					},
				},
			},
		},
	}

	body := new(bytes.Buffer)
	writer := multipart.NewWriter(body)

	part, _ := writer.CreateFormFile("taxFile", "test.csv")

	f, err := os.OpenFile("./data/sample.csv", os.O_RDONLY, 0644)
	if err != nil {
		t.Fatal(err)
	}

	defer f.Close()

	w, err := io.Copy(part, f)
	if err != nil || w == 0 {
		t.Fatal("failed to copy file", err)
	}

	t.Logf("body: %v", body.String())

	req := httptest.NewRequest(http.MethodPost, "/tax/calculations/upload-csv", body)
	req.Header.Set(echo.HeaderContentType, writer.FormDataContentType())
	rec := httptest.NewRecorder()

	writer.Close()

	e := echo.New()
	c := e.NewContext(req, rec)
	c.SetPath("/tax/calculations/upload-csv")

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
	err = h.CalculateTaxFromTaxFile(c)
	if err != nil {
		t.Errorf("unable to calculate tax from file: %v", err)
	}

	if rec.Code != http.StatusOK {
		t.Errorf("invalid http status: got %v want %v",
			rec.Code, http.StatusOK)
	}

	var res struct {
		Taxes []TaxCalculationResponse `json:"taxes"`
	}

	err = json.Unmarshal(rec.Body.Bytes(), &res)
	if err != nil {
		t.Errorf("unable to unmarshal response: %v", err)
	}

	log.Println(rec.Body.String())
	if !reflect.DeepEqual(res, expectedResult) {
		t.Errorf("invalid result: got %v want %v",
			res, expectedResult)
	}

}
