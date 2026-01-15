package handler

import (
	"net/http"
	"os"
	"path/filepath"

	"github.com/labstack/echo/v4"
)

type TagHandler struct {
}

func NewTagHandler() *TagHandler {
	return &TagHandler{}
}

type tagsResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func (h *TagHandler) GetTags(c echo.Context) error {
	name := c.Param("name")

	tagFiles, err := os.ReadDir(filepath.Join(tagDir, name))
	if err != nil {
		return c.String(http.StatusInternalServerError, "failed to read tag dir")
	}

	tags := make([]string, 0, len(tagFiles))
	for _, file := range tagFiles {
		tags = append(tags, file.Name())
	}

	res := &tagsResponse{
		Name: name,
		Tags: tags,
	}
	return c.JSON(http.StatusOK, res)
}
