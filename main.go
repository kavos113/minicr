package main

import (
	"log"
	"net/http"

	"github.com/kavos113/minicr/handler"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	e := echo.New()
	e.Use(middleware.RequestLoggerWithConfig(middleware.RequestLoggerConfig{
		LogStatus: true,
		LogMethod: true,
		LogURI: true,
		LogError: true,
		LogHeaders: []string{
			"Content-Type",
			"Content-Length",
			"Content-Range",
		},
		LogValuesFunc: func(c echo.Context, v middleware.RequestLoggerValues) error {
			log.Printf("Status: %d, Method: %s, URI: %s, Headers: %+v, Error: %+v", v.Status, v.Method, v.URI, v.Headers, v.Error)
			return nil
		},
	}))

	if err := handler.InitDirs(); err != nil {
		e.Logger.Fatal("Failed to initialize directories: ", err)
	}

	ph := handler.NewBlobUploadHandler()
	mh := handler.NewManifestHandler()

	e.GET("/v2/", baseHandler)
	e.POST("/v2/:name/blobs/uploads/", ph.PostBlobUploads)
	e.PUT("/v2/:name/blobs/uploads/:reference", ph.PutBlobUpload)
	e.PATCH("/v2/:name/blobs/uploads/:reference", ph.PatchBlobUpload)
	e.PUT("/v2/:name/manifests/:reference", mh.PutManifests)

	e.Logger.Fatal(e.Start(":8080"))
}

func baseHandler(c echo.Context) error {
	c.Response().Header().Set("Content-Type", "application/json")
	return c.JSON(http.StatusOK, map[string]string{})
}
