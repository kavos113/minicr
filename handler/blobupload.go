package handler

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/opencontainers/go-digest"
)

type BlobUploadHandler struct {
}

func NewBlobUploadHandler() *BlobUploadHandler {
	return &BlobUploadHandler{}
}

func (h *BlobUploadHandler) PostBlobUploads(c echo.Context) error {
	name := c.Param("name")

	id, err := uuid.NewV7()
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to generate upload ID")
	}

	c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("/v2/%s/blobs/uploads/%s", name, id.String()))
	c.Response().Header().Set("Docker-Upload-UUID", id.String())

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

	tmpPath := filepath.Join(uploadDir, reference)
	tmpFile, err := os.OpenFile(tmpPath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("failed to open upload file %s: %+v", tmpPath, err)
		return c.String(http.StatusInternalServerError, "failed to open upload file")
	}
	defer tmpFile.Close()

	_, err = io.Copy(tmpFile, c.Request().Body)
	if err != nil {
		log.Printf("failed to save file content to %s: %+v", tmpPath, err)
		return c.String(http.StatusInternalServerError, "failed to save file content")
	}

	_, err = tmpFile.Seek(0, io.SeekStart)
	if err != nil {
		log.Printf("failed to seek start of file %s: %+v", tmpPath, err)
		return c.String(http.StatusInternalServerError, "failed to seek file")
	}

	verifier := d.Verifier()
	_, err = io.Copy(verifier, tmpFile)
	if err != nil {
		log.Printf("failed to verify digest for file %s: %+v", tmpPath, err)
		return c.String(http.StatusInternalServerError, "failed to verify")
	}
	if !verifier.Verified() {
		log.Printf("not verified digest: %s", dstr)
		return c.String(http.StatusBadRequest, "uploaded data does not match the provided digest")
	}

	if err = os.MkdirAll(filepath.Join(blobDir, name), 0755); err != nil {
		return c.String(http.StatusInternalServerError, "failed to create blob dir")
	}

	blobPath := filepath.Join(blobDir, name, d.String())
	if err := os.Rename(tmpPath, blobPath); err != nil {
		log.Printf("failed to store blob %s: %+v", blobPath, err)
		return c.String(http.StatusInternalServerError, "failed to store blob")
	}

	c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("/v2/%s/blobs/%s", name, d.String()))
	c.Response().Header().Set("Docker-Content-Digest", d.String())

	return c.NoContent(http.StatusCreated)
}

func (h *BlobUploadHandler) PatchBlobUpload(c echo.Context) error {
	name := c.Param("name")
	reference := c.Param("reference")

	// TODO validation by Content-Length, Content-Type

	tmpPath := filepath.Join(uploadDir, reference)
	f, err := os.OpenFile(tmpPath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		log.Printf("failed to open upload file %s: %+v", tmpPath, err)
		return c.String(http.StatusInternalServerError, "failed to open upload file")
	}
	defer f.Close()

	_, err = io.Copy(f, c.Request().Body)
	if err != nil {
		log.Printf("failed to save file content to %s: %+v", tmpPath, err)
		return c.String(http.StatusInternalServerError, "failed to save file content")
	}

	stat, err := os.Stat(tmpPath)
	if err != nil {
		log.Printf("failed to get tmp file stat %s: %+v", tmpPath, err)
		return c.String(http.StatusInternalServerError, "failed to get tmp file stat")
	}

	c.Response().Header().Set(echo.HeaderLocation, fmt.Sprintf("/v2/%s/blobs/uploads/%s", name, reference))
	c.Response().Header().Set("Range", fmt.Sprintf("0-%d", stat.Size()))

	return c.NoContent(http.StatusAccepted)
}
