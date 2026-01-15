package handler

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/kavos113/minicr/schema"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
)

type ManifestHandler struct {
}

func NewManifestHandler() *ManifestHandler {
	return &ManifestHandler{}
}

func (h *ManifestHandler) PutManifests(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return c.String(http.StatusBadRequest, "blank repo name")
	}

	// TODO: tag/digest
	tag := c.Param("reference")
	if tag == "" {
		return c.String(http.StatusBadRequest, "blank reference")
	}

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

		blobPath := filepath.Join(blobDir, desc.Digest.String())
		if _, err := os.Stat(blobPath); os.IsNotExist(err) {
			// TODO: error code specs
			return c.String(http.StatusBadRequest, "blob unknown")
		}
	}

	d := digest.FromBytes(payload)

	manifestPath := filepath.Join(blobDir, d.String())
	err = os.WriteFile(manifestPath, payload, 0644)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to save manifest file")
	}

	c.Response().Header().Set("Location", fmt.Sprintf("/v2/%s/manifests/%s/", name, d.String()))
	c.Response().Header().Set("Docker-Content-Digest", d.String())

	return c.NoContent(http.StatusCreated)
}
