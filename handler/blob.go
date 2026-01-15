package handler

import (
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
)

type BlobHandler struct {
}

func NewBlobHandler() *BlobHandler {
	return &BlobHandler{}
}

func (b *BlobHandler) GetBlobs(c echo.Context) error {
	name := c.Param("name")
	dstr := c.Param("digest")

	blobPath := filepath.Join(blobDir, name, dstr)
	s, err := os.Stat(blobPath)
	if err != nil {
		if os.IsNotExist(err) {
			return c.NoContent(http.StatusNotFound)
		}
		return c.String(http.StatusInternalServerError, "failed to stat")
	}

	f, err := os.Open(blobPath)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to open file")
	}
	defer f.Close()

	d, err := digest.FromReader(f)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to create digest")
	}
	_, err = f.Seek(0, io.SeekStart)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to seek blob file")
	}

	_, err = io.Copy(c.Response().Writer, f)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to copy response")
	}

	c.Response().Header().Set("Docker-Content-Digest", d.String())
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEOctetStream)
	c.Response().Header().Set(echo.HeaderContentLength, strconv.FormatInt(s.Size(), 10))

	log.Printf("response: %+v\n", *c.Response())

	return c.NoContent(http.StatusOK)
}

func (b *BlobHandler) HeadBlobs(c echo.Context) error {
	name := c.Param("name")
	dstr := c.Param("digest")

	blobPath := filepath.Join(blobDir, name, dstr)
	s, err := os.Stat(blobPath);
	if err != nil {
		if os.IsNotExist(err) {
			return c.NoContent(http.StatusNotFound)
		}
		return c.String(http.StatusInternalServerError, "failed to stat")
	}

	f, err := os.Open(blobPath)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to open file")
	}
	defer f.Close()

	d, err := digest.FromReader(f)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to create digest")
	}

	c.Response().Header().Set("Docker-Content-Digest", d.String())
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEOctetStream)
	c.Response().Header().Set(echo.HeaderContentLength, strconv.FormatInt(s.Size(), 10))

	return c.NoContent(http.StatusOK)
}
