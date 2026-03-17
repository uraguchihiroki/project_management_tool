package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type CommentHandler struct {
	commentRepo repository.CommentRepository
}

func NewCommentHandler(commentRepo repository.CommentRepository) *CommentHandler {
	return &CommentHandler{commentRepo: commentRepo}
}

func (h *CommentHandler) List(c echo.Context) error {
	issueID, err := uuid.Parse(c.Param("issueId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid issue id")
	}
	comments, err := h.commentRepo.FindByIssue(issueID)
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
	comment := &model.Comment{
		ID:        uuid.New(),
		IssueID:   issueID,
		AuthorID:  authorID,
		Body:      req.Body,
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
	if err := h.commentRepo.Create(comment); err != nil {
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
	_ = id
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "updated"})
}

func (h *CommentHandler) Delete(c echo.Context) error {
	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid comment id")
	}
	if err := h.commentRepo.Delete(id); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"message": "deleted"})
}
