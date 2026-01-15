package handler

import (
	"errors"
	"net/http"
	"strconv"

	"github.com/kavos113/minicr/storage"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
)

type BlobHandler struct {
	storage storage.Storage
}

func NewBlobHandler(s storage.Storage) *BlobHandler {
	return &BlobHandler{storage: s}
}

func (h *BlobHandler) GetBlobs(c echo.Context) error {
	name := c.Param("name")
	dstr := c.Param("digest")

	d, err := digest.Parse(dstr)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	size, err := h.storage.ReadBlobToWriter(name, d, c.Response().Writer)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		if errors.Is(err, storage.ErrNotVerified) {
			return c.NoContent(http.StatusBadRequest)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	c.Response().Header().Set("Docker-Content-Digest", d.String())
	c.Response().Header().Set(echo.HeaderContentType, echo.MIMEOctetStream)
	c.Response().Header().Set(echo.HeaderContentLength, strconv.FormatInt(size, 10))

	return c.NoContent(http.StatusOK)
}
