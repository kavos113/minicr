package handler

import (
	"errors"
	"fmt"
	"log"
	"net/http"

	"github.com/google/uuid"
	"github.com/kavos113/minicr/storage"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
)

type BlobUploadHandler struct {
	storage storage.Storage
}

func NewBlobUploadHandler(s storage.Storage) *BlobUploadHandler {
	return &BlobUploadHandler{storage: s}
}

func (h *BlobUploadHandler) PostBlobUploads(c echo.Context) error {
	name := c.Param("name")

	id, err := uuid.NewV7()
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to generate upload ID")
	}
	c.Response().Header().Set("Docker-Upload-UUID", id.String())

	dstr := c.QueryParam("digest")
	if dstr != "" {
		// monolithic upload
		d, err := digest.Parse(dstr)
		if err != nil {
			log.Printf("cannot parse digest %s: %+v", dstr, err)
			return c.String(http.StatusBadRequest, "invalid digest format")
		}

		_, err = h.storage.UploadBlob(id.String(), c.Request().Body)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}

		err = h.storage.CommitBlob(name, id.String(), d)
		if err != nil {
			if errors.Is(err, storage.ErrNotVerified) {
				return c.NoContent(http.StatusBadRequest)
			}
			return c.NoContent(http.StatusInternalServerError)
		}

		c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("/v2/%s/blobs/%s", name, d.String()))
		return c.NoContent(http.StatusCreated)
	}

	c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("/v2/%s/blobs/uploads/%s", name, id.String()))

	return c.NoContent(http.StatusAccepted)
}

func (h *BlobUploadHandler) PatchBlobUpload(c echo.Context) error {
	name := c.Param("name")
	reference := c.Param("reference")

	// TODO validation by Content-Length, Content-Type
	size, err := h.storage.UploadBlob(reference, c.Request().Body)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("/v2/%s/blobs/uploads/%s", name, reference))
	c.Response().Header().Set("Range", fmt.Sprintf("0-%d", size))

	return c.NoContent(http.StatusAccepted)
}

func (h *BlobUploadHandler) PutBlobUpload(c echo.Context) error {
	name := c.Param("name")
	reference := c.Param("reference")

	dstr := c.QueryParam("digest")
	if dstr == "" {
		return c.String(http.StatusBadRequest, "digest query parameter is required")
	}

	d, err := digest.Parse(dstr)
	if err != nil {
		log.Printf("cannot parse digest %s: %+v", dstr, err)
		return c.String(http.StatusBadRequest, "invalid digest format")
	}

	_, err = h.storage.UploadBlob(reference, c.Request().Body)
	if err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	err = h.storage.CommitBlob(name, reference, d)
	if err != nil {
		if errors.Is(err, storage.ErrNotVerified) {
			return c.NoContent(http.StatusBadRequest)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("/v2/%s/blobs/%s", name, d.String()))
	c.Response().Header().Set("Docker-Content-Digest", d.String())

	return c.NoContent(http.StatusCreated)
}
