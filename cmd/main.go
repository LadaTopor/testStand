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

	// Start server
	e.Logger.Fatal(e.Start(":1323"))
}
