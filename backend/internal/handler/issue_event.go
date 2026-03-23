package handler

import (
	"net/http"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
)

type IssueEventHandler struct {
	issueRepo repository.IssueRepository
	eventRepo repository.IssueEventRepository
}

func NewIssueEventHandler(issueRepo repository.IssueRepository, eventRepo repository.IssueEventRepository) *IssueEventHandler {
	return &IssueEventHandler{issueRepo: issueRepo, eventRepo: eventRepo}
}

// GET /api/v1/issues/:issueId/events
func (h *IssueEventHandler) ListByIssue(c echo.Context) error {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return authErr
	}
	issueID, err := uuid.Parse(c.Param("issueId"))
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid issue id")
	}
	issue, err := h.issueRepo.FindByID(issueID)
	if err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "issue not found")
	}
	if !isSuperAdmin && (orgScope == nil || issue.OrganizationID != *orgScope) {
		return echo.NewHTTPError(http.StatusNotFound, "issue not found")
	}
	events, err := h.eventRepo.ListByIssueID(issueID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": events})
}

// GET /api/v1/organizations/:orgId/issue-events
func (h *IssueEventHandler) ListByOrganization(c echo.Context) error {
	orgID, _, authErr := requireOrgParam(c, "orgId")
	if authErr != nil {
		return authErr
	}
	f := repository.IssueEventFilters{}
	if v := c.QueryParam("event_type"); v != "" {
		f.EventType = &v
	}
	if v := c.QueryParam("from_occurred_at"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid from_occurred_at (use RFC3339)")
		}
		f.FromOccurredAt = &t
	}
	if v := c.QueryParam("to_occurred_at"); v != "" {
		t, err := time.Parse(time.RFC3339, v)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid to_occurred_at (use RFC3339)")
		}
		f.ToOccurredAt = &t
	}
	if v := c.QueryParam("actor_id"); v != "" {
		aid, err := uuid.Parse(v)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid actor_id")
		}
		f.ActorID = &aid
	}
	if v := c.QueryParam("issue_id"); v != "" {
		iid, err := uuid.Parse(v)
		if err != nil {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid issue_id")
		}
		f.IssueID = &iid
	}
	events, err := h.eventRepo.ListByOrganization(orgID, f)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": events})
}
