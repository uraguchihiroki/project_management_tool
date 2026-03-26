package handler

import (
	"net/http"
	"strconv"
	"time"

	"github.com/google/uuid"
	"github.com/labstack/echo/v4"
	"github.com/uraguchihiroki/project_management_tool/internal/model"
	"github.com/uraguchihiroki/project_management_tool/internal/pkg/keygen"
	"github.com/uraguchihiroki/project_management_tool/internal/repository"
	"github.com/uraguchihiroki/project_management_tool/internal/service"
)

type WorkflowTransitionHandler struct {
	workflowService service.WorkflowService
	statusService   service.StatusService
	transitionRepo  repository.WorkflowTransitionRepository
}

func NewWorkflowTransitionHandler(
	workflowService service.WorkflowService,
	statusService service.StatusService,
	transitionRepo repository.WorkflowTransitionRepository,
) *WorkflowTransitionHandler {
	return &WorkflowTransitionHandler{
		workflowService: workflowService,
		statusService:   statusService,
		transitionRepo:  transitionRepo,
	}
}

func (h *WorkflowTransitionHandler) authorizeWorkflowAccess(c echo.Context) (uint, error) {
	orgScope, isSuperAdmin, authErr := requireClaims(c)
	if authErr != nil {
		return 0, authErr
	}
	id, err := strconv.ParseUint(c.Param("id"), 10, 64)
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusBadRequest, "invalid workflow id")
	}
	workflow, err := h.workflowService.GetWorkflow(uint(id))
	if err != nil {
		return 0, echo.NewHTTPError(http.StatusNotFound, "workflow not found")
	}
	if !isSuperAdmin && (orgScope == nil || workflow.OrganizationID != *orgScope) {
		return 0, echo.NewHTTPError(http.StatusNotFound, "workflow not found")
	}
	return uint(id), nil
}

// GET /api/v1/workflows/:id/transitions
func (h *WorkflowTransitionHandler) ListByWorkflow(c echo.Context) error {
	wfID, err := h.authorizeWorkflowAccess(c)
	if err != nil {
		return err
	}
	rows, err := h.transitionRepo.FindByWorkflowID(wfID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": rows})
}

// POST /api/v1/workflows/:id/transitions
func (h *WorkflowTransitionHandler) CreateForWorkflow(c echo.Context) error {
	wfID, err := h.authorizeWorkflowAccess(c)
	if err != nil {
		return err
	}
	type Request struct {
		FromStatusID string `json:"from_status_id"`
		ToStatusID   string `json:"to_status_id"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	fromID, err := uuid.Parse(req.FromStatusID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid from_status_id")
	}
	toID, err := uuid.Parse(req.ToStatusID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid to_status_id")
	}
	fromSt, err := h.statusService.Get(fromID)
	if err != nil || fromSt.WorkflowID != wfID {
		return echo.NewHTTPError(http.StatusBadRequest, "from_status_id is not in this workflow")
	}
	toSt, err := h.statusService.Get(toID)
	if err != nil || toSt.WorkflowID != wfID {
		return echo.NewHTTPError(http.StatusBadRequest, "to_status_id is not in this workflow")
	}
	if fromID == toID {
		return echo.NewHTTPError(http.StatusBadRequest, "from and to status must differ")
	}
	if h.transitionRepo.Exists(wfID, fromID, toID) {
		return echo.NewHTTPError(http.StatusBadRequest, "transition already exists")
	}
	nextDO, err := h.transitionRepo.MaxDisplayOrder(wfID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	row := &model.WorkflowTransition{
		Key:          keygen.UUIDKey(uuid.New()),
		WorkflowID:   wfID,
		FromStatusID: fromID,
		ToStatusID:   toID,
		DisplayOrder: nextDO + 1,
		CreatedAt:    time.Now(),
	}
	if err := h.transitionRepo.Create(row); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	created, err := h.transitionRepo.FindByID(row.ID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusCreated, map[string]interface{}{"data": created})
}

// DELETE /api/v1/workflows/:id/transitions/:transitionId
func (h *WorkflowTransitionHandler) Delete(c echo.Context) error {
	wfID, err := h.authorizeWorkflowAccess(c)
	if err != nil {
		return err
	}
	id, err := strconv.ParseUint(c.Param("transitionId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid transition id")
	}
	row, err := h.transitionRepo.FindByID(uint(id))
	if err != nil || row.WorkflowID != wfID {
		return echo.NewHTTPError(http.StatusNotFound, "transition not found")
	}
	if err := h.transitionRepo.DeleteByID(uint(id)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}

// PUT /api/v1/workflows/:id/transitions/:transitionId
func (h *WorkflowTransitionHandler) Update(c echo.Context) error {
	wfID, err := h.authorizeWorkflowAccess(c)
	if err != nil {
		return err
	}
	tid, err := strconv.ParseUint(c.Param("transitionId"), 10, 64)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid transition id")
	}
	row, err := h.transitionRepo.FindByID(uint(tid))
	if err != nil || row.WorkflowID != wfID {
		return echo.NewHTTPError(http.StatusNotFound, "transition not found")
	}
	type Request struct {
		FromStatusID string `json:"from_status_id"`
		ToStatusID   string `json:"to_status_id"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	fromID, err := uuid.Parse(req.FromStatusID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid from_status_id")
	}
	toID, err := uuid.Parse(req.ToStatusID)
	if err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid to_status_id")
	}
	if fromID == toID {
		return echo.NewHTTPError(http.StatusBadRequest, "from and to status must differ")
	}
	fromSt, err := h.statusService.Get(fromID)
	if err != nil || fromSt.WorkflowID != wfID {
		return echo.NewHTTPError(http.StatusBadRequest, "from_status_id is not in this workflow")
	}
	toSt, err := h.statusService.Get(toID)
	if err != nil || toSt.WorkflowID != wfID {
		return echo.NewHTTPError(http.StatusBadRequest, "to_status_id is not in this workflow")
	}
	if h.transitionRepo.ExistsOtherThan(wfID, fromID, toID, uint(tid)) {
		return echo.NewHTTPError(http.StatusBadRequest, "transition already exists")
	}
	if err := h.transitionRepo.Update(fromID, toID, uint(tid)); err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	updated, err := h.transitionRepo.FindByID(uint(tid))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	return c.JSON(http.StatusOK, map[string]interface{}{"data": updated})
}

// PUT /api/v1/workflows/:id/transitions/reorder
func (h *WorkflowTransitionHandler) ReorderForWorkflow(c echo.Context) error {
	wfID, err := h.authorizeWorkflowAccess(c)
	if err != nil {
		return err
	}
	type Request struct {
		TransitionIDs []uint `json:"transition_ids"`
	}
	var req Request
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	rows, err := h.transitionRepo.FindByWorkflowID(wfID)
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, err.Error())
	}
	if len(req.TransitionIDs) != len(rows) {
		return echo.NewHTTPError(http.StatusBadRequest, "transition_ids must list all transitions for this workflow")
	}
	seen := make(map[uint]struct{}, len(rows))
	for _, r := range rows {
		seen[r.ID] = struct{}{}
	}
	for _, id := range req.TransitionIDs {
		if _, ok := seen[id]; !ok {
			return echo.NewHTTPError(http.StatusBadRequest, "invalid transition id")
		}
		delete(seen, id)
	}
	if len(seen) != 0 {
		return echo.NewHTTPError(http.StatusBadRequest, "transition_ids must be a permutation of workflow transitions")
	}
	if err := h.transitionRepo.ReorderWorkflow(wfID, req.TransitionIDs); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, err.Error())
	}
	return c.NoContent(http.StatusNoContent)
}
