package handler

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"

	"github.com/kavos113/minicr/schema"
	"github.com/kavos113/minicr/storage"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
)

type ManifestHandler struct {
	storage storage.Storage
}

func NewManifestHandler(s storage.Storage) *ManifestHandler {
	return &ManifestHandler{storage: s}
}

func (h *ManifestHandler) PutManifests(c echo.Context) error {
	name := c.Param("name")
	ref := c.Param("reference")
	istag := isTag(ref)

	payload, err := io.ReadAll(c.Request().Body)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to read manifest")
	}

	var m schema.Manifest
	if err := json.Unmarshal(payload, &m); err != nil {
		return c.String(http.StatusBadRequest, "invalid manifest")
	}

	for _, desc := range append(m.Layers, m.Config) {
		err = desc.Digest.Validate()
		if err != nil {
			return c.String(http.StatusBadRequest, "invalid digest")
		}

		exist, err := h.storage.IsExistBlob(name, desc.Digest)
		if err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
		if !exist {
			return c.String(http.StatusBadRequest, "unknown blob layer")
		}
	}

	d := digest.FromBytes(payload)

	if err := h.storage.SaveBlob(name, d, payload); err != nil {
		return c.NoContent(http.StatusInternalServerError)
	}

	if istag {
		if err := h.storage.SaveTag(name, d, ref); err != nil {
			return c.NoContent(http.StatusInternalServerError)
		}
	}

	c.Response().Header().Set("Location", fmt.Sprintf("/v2/%s/manifests/%s/", name, d.String()))
	c.Response().Header().Set("Docker-Content-Digest", d.String())

	return c.NoContent(http.StatusCreated)
}

func (h *ManifestHandler) GetManifests(c echo.Context) error {
	name := c.Param("name")
	ref := c.Param("reference")
	istag := isTag(ref)

	dstr := ref
	if istag {
		tag, err := h.storage.ReadTag(name, ref)
		if err != nil {
			if errors.Is(err, storage.ErrNotFound) {
				return c.NoContent(http.StatusNotFound)
			}
			return c.NoContent(http.StatusInternalServerError)
		}
		dstr = tag
	}

	d, err := digest.Parse(dstr)
	if err != nil {
		return c.NoContent(http.StatusBadRequest)
	}

	rawManifest, err := h.storage.ReadBlob(name, d)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	var m schema.Manifest
	if err = json.Unmarshal(rawManifest, &m); err != nil {
		return c.String(http.StatusInternalServerError, "failed to parse json manifest")
	}

	c.Response().Header().Set(echo.HeaderContentType, m.MediaType)
	c.Response().Header().Set("Docker-Content-Digest", d.String())

	return c.JSON(http.StatusOK, m)
}

func isTag(reference string) bool {
	_, err := digest.Parse(reference)
	return err != nil
}
