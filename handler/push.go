package handler

import (
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
)

var (
	rootPath  = "./data"
	uploadDir = filepath.Join(rootPath, "uploads")
	blobDir   = filepath.Join(rootPath, "blobs")
)

func InitDirs() error {
	dirs := []string{rootPath, uploadDir, blobDir}
	for _, dir := range dirs {
		if err := os.MkdirAll(dir, 0755); err != nil {
			return err
		}
	}
	return nil
}

type PushHandler struct {
}

func NewPushHandler() *PushHandler {
	return &PushHandler{}
}

func (h *PushHandler) PostBlobUploads(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return c.String(http.StatusBadRequest, "repository name is required")
	}

	id, err := uuid.NewV7()
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to generate upload ID")
	}

	c.Response().Header().Set("Location", fmt.Sprintf("/v2/%s/blobs/uploads/%s", name, id.String()))
	c.Response().Header().Set("Docker-Upload-UUID", id.String())

	return c.NoContent(http.StatusAccepted)
}

func (h *PushHandler) PutBlobUpload(c echo.Context) error {
	name := c.Param("name")
	if name == "" {
		return c.String(http.StatusBadRequest, "repository name is required")
	}

	reference := c.Param("reference")
	if reference == "" {
		return c.String(http.StatusBadRequest, "blob reference is required")
	}

	dstr := c.QueryParam("digest")
	if dstr == "" {
		return c.String(http.StatusBadRequest, "digest query parameter is required")
	}

	d, err := digest.Parse(dstr)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid digest format")
	}

	tmpPath := filepath.Join(uploadDir, reference)
	tmpFile, err := os.Create(tmpPath)
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to create temporary file for upload")
	}
	defer tmpFile.Close()

	verifier := d.Verifier()
	writer := io.MultiWriter(tmpFile, verifier)

	if _, err := io.Copy(writer, c.Request().Body); err != nil {
		return c.String(http.StatusInternalServerError, "failed to read upload data")
	}

	if !verifier.Verified() {
		return c.String(http.StatusBadRequest, "uploaded data does not match the provided digest")
	}

	blobPath := filepath.Join(blobDir, d.String())
	if err := os.Rename(tmpPath, blobPath); err != nil {
		return c.String(http.StatusInternalServerError, "failed to store blob")
	}

	c.Response().Header().Set("Location", fmt.Sprintf("/v2/%s/blobs/%s", name, d.String()))
	c.Response().Header().Set("Docker-Content-Digest", d.String())

	return c.NoContent(http.StatusCreated)
}
