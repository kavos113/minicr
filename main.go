package main

import (
	"net/http"

	"github.com/kavos113/minicr/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.RequestLogger())

	if err := handler.InitDirs(); err != nil {
		e.Logger.Fatal("Failed to initialize directories: ", err)
	}

	ph := handler.NewPushHandler()
	mh := handler.NewManifestHandler()

	e.GET("/v2/", baseHandler)
	e.POST("/v2/:name/blobs/uploads/", ph.PostBlobUploads)
	e.PUT("/v2/:name/blobs/uploads/:reference", ph.PutBlobUpload)
	e.PUT("/v2/:name/manifests/:reference", mh.PutManifests)

	e.Logger.Fatal(e.Start(":8080"))
}

func baseHandler(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json")
	return c.JSON(http.StatusOK, map[string]string{})
}
