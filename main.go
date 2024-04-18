package main

import (
	"net/http"

	"github.com/bytesbanana/assessment-tax/tax"
	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Go Bootcamp!")
	})

	taxHandler := tax.NewHandler()
	e.POST("/tax/calculations", taxHandler.CalculateTax)

	e.Logger.Fatal(e.Start(":1323"))
}
