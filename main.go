package main

import (
	"net/http"

	"github.com/labstack/echo/v4"
)

func main() {
	e := echo.New()

	e.GET("/v2/", baseHandler)

	e.Logger.Fatal(e.Start(":8080"))
}

func baseHandler(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json")
	return c.JSON(http.StatusOK, map[string]string{})
}
