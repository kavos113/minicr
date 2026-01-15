package handler

import (
	"errors"
	"net/http"

	"github.com/kavos113/minicr/storage"
	"github.com/labstack/echo/v4"
)

type TagHandler struct {
	storage storage.Storage
}

func NewTagHandler(s storage.Storage) *TagHandler {
	return &TagHandler{storage: s}
}

type tagsResponse struct {
	Name string   `json:"name"`
	Tags []string `json:"tags"`
}

func (h *TagHandler) GetTags(c echo.Context) error {
	name := c.Param("name")

	tags, err := h.storage.GetTagList(name)
	if err != nil {
		if errors.Is(err, storage.ErrNotFound) {
			return c.NoContent(http.StatusNotFound)
		}
		return c.NoContent(http.StatusInternalServerError)
	}

	res := &tagsResponse{
		Name: name,
		Tags: tags,
	}
	return c.JSON(http.StatusOK, res)
}
