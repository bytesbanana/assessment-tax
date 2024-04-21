package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"github.com/bytesbanana/assessment-tax/admin"
	"github.com/bytesbanana/assessment-tax/postgres"
	"github.com/bytesbanana/assessment-tax/tax"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func basicAuthMiddleware() echo.MiddlewareFunc {
	return middleware.BasicAuth(func(username, password string, c echo.Context) (bool, error) {

		if username == os.Getenv("ADMIN_USERNAME") && password == os.Getenv("ADMIN_PASSWORD") {
			return true, nil
		}

		return false, nil
	})
}

func main() {
	port, err := strconv.Atoi(os.Getenv("PORT"))
	if err != nil {
		log.Fatalf("invalid port: %v", err)
	}

	p, err := postgres.New()
	if err != nil {
		panic(err)
	}

	e := echo.New()
	e.GET("/", func(c echo.Context) error {
		return c.String(http.StatusOK, "Hello, Go Bootcamp!")
	})

	taxHandler := tax.New(p)
	e.POST("/tax/calculations", taxHandler.CalculateTax)

	adminHandler := admin.New(p)
	adminGroup := e.Group("/admin")
	adminGroup.Use(basicAuthMiddleware())
	adminGroup.POST("/deductions/k-receipt", adminHandler.SetPersonalDeductionsConfig)

	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	go func() {
		if err := e.Start(fmt.Sprintf(":%d", port)); err != nil && err != http.ErrServerClosed {
			e.Logger.Fatal("shutting down the server")
		}
	}()

	<-ctx.Done()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := e.Shutdown(ctx); err != nil {
		e.Logger.Fatal(err)
	}

}
