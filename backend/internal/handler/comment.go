package handler

import (
	"net/http"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
)

type CommentHandler struct {
	commentService service.CommentService
}

func NewCommentHandler(commentService service.CommentService) *CommentHandler {
	return &CommentHandler{commentService: commentService}
}

func (h *CommentHandler) List(c echo.Context) error {
	issueID, err := uuid.Parse(c.Param("issueId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid issue id")
	}
	comments, err := h.commentService.List(issueID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": comments})
}

func (h *CommentHandler) Create(c echo.Context) error {
	issueID, err := uuid.Parse(c.Param("issueId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid issue id")
	}
	type Request struct {
		AuthorID string `json:"author_id" validate:"required,uuid"`
		Body     string `json:"body" validate:"required"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	authorID, err := uuid.Parse(req.AuthorID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid author_id")
	}
	comment, err := h.commentService.Create(issueID, authorID, req.Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": comment})
}

func (h *CommentHandler) Update(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid comment id")
	}
	type Request struct {
		Body string `json:"body" validate:"required"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	comment, err := h.commentService.Update(id, req.Body)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "comment not found")
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": comment})
}

func (h *CommentHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid comment id")
	}
	if err := h.commentService.Delete(id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "deleted"})
}
