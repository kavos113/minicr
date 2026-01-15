package handler

import (
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
)

type BlobHandler struct {
}

func NewBlobHandler() *BlobHandler {
	return &BlobHandler{}
}

func (b *BlobHandler) GetBlobs(c echo.Context) error {
	dstr := c.Param("digest")

	blobPath := filepath.Join(blobDir, dstr)
	if _, err := os.Stat(blobPath); err != nil {
		if os.IsNotExist(err) {
			return c.NoContent(http.StatusNotFound)
		}
		return c.String(http.StatusInternalServerError, "failed to stat")
	}

	f, err := os.Open(blobPath)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to open file")
	}

	d, err := digest.FromReader(f)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to create digest")
	}

	_, err = io.Copy(c.Response().Writer, f)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to copy response")
	}

	c.Response().Header().Set("Docker-Content-Digest", d.String())
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEOctetStream)

	return c.NoContent(http.StatusOK)
}
