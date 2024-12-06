package main

import (
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"testStand/internal/service"
)

func main() {
	// Echo instance
	e := echo.New()

	// Middleware
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	db := PostgresConnection()

	svc := service.NewService(db)

	// Routes
	e.POST("/payout", svc.CreatePayoutTransaction)
	e.POST("/payment", svc.CreatePaymentTransaction)

	e.POST("/callback/:acquirer", svc.CallbackHandler)

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
