package task_controller

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"taskflow/delivery/http/common"
	domain_error "taskflow/internal/domain/errors"
	domain_task "taskflow/internal/domain/task"
	"taskflow/utils/validator"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Controller struct {
	tasks domain_task.UseCase
}

func New(tasks domain_task.UseCase) *Controller {
	return &Controller{tasks: tasks}
}

// List godoc
// GET /v1/projects/{projectID}/tasks
func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	projectID, err := uuid.Parse(chi.URLParam(r, "projectID"))
	if err != nil {
		common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid project id", err))
		return
	}

	filter := domain_task.ListFilter{}
	if s := r.URL.Query().Get("status"); s != "" {
		status := domain_task.Status(s)
		filter.Status = &status
	}
	if a := r.URL.Query().Get("assignee"); a != "" {
		aid, err := uuid.Parse(a)
		if err != nil {
			common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid assignee id", err))
			return
		}
		filter.AssigneeID = &aid
	}

	page := 1
	pageSize := 20
	if p := r.URL.Query().Get("page"); p != "" {
		if v, err := strconv.Atoi(p); err == nil && v > 0 {
			page = v
		}
	}
	if ps := r.URL.Query().Get("page_size"); ps != "" {
		if v, err := strconv.Atoi(ps); err == nil && v > 0 && v <= 100 {
			pageSize = v
		}
	}

	tasks, err := c.tasks.List(r.Context(), projectID, filter, domain_task.Pagination{Page: page, PageSize: pageSize})
	if err != nil {
		common.SendAppError(w, err)
		return
	}

	if tasks == nil {
		tasks = []*domain_task.Task{}
	}
	common.SendJSON(w, http.StatusOK, map[string]any{"tasks": tasks})
}

// Create godoc
// POST /v1/projects/{projectID}/tasks
func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	callerID := r.Context().Value(common.UserIDKey).(uuid.UUID)

	projectID, err := uuid.Parse(chi.URLParam(r, "projectID"))
	if err != nil {
		common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid project id", err))
		return
	}

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "", err))
		return
	}
	if err := validator.Validate(r.Context(), req); err != nil {
		common.SendAppError(w, err)
		return
	}

	input := domain_task.CreateInput{
		ProjectID: projectID,
		Title:     req.Title,
		Priority:  req.Priority,
	}

	if req.AssigneeID != nil {
		aid, err := uuid.Parse(*req.AssigneeID)
		if err != nil {
			common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid assignee_id", err))
			return
		}
		input.AssigneeID = &aid
	}

	if req.DueDate != nil {
		t, err := time.Parse(time.DateOnly, *req.DueDate)
		if err != nil {
			common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid due_date, expected YYYY-MM-DD", err))
			return
		}
		input.DueDate = &t
	}

	task, err := c.tasks.Create(r.Context(), callerID, input)
	if err != nil {
		common.SendAppError(w, err)
		return
	}

	common.SendJSON(w, http.StatusCreated, task)
}

// Update godoc
// PATCH /v1/tasks/{id}
func (c *Controller) Update(w http.ResponseWriter, r *http.Request) {
	callerID := r.Context().Value(common.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid task id", err))
		return
	}

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "", err))
		return
	}

	input := domain_task.UpdateInput{
		Title:    req.Title,
		Status:   req.Status,
		Priority: req.Priority,
	}

	if req.AssigneeID != nil {
		aid, err := uuid.Parse(*req.AssigneeID)
		if err != nil {
			common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid assignee_id", err))
			return
		}
		input.AssigneeID = &aid
	}

	if req.DueDate != nil {
		t, err := time.Parse(time.DateOnly, *req.DueDate)
		if err != nil {
			common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid due_date, expected YYYY-MM-DD", err))
			return
		}
		input.DueDate = &t
	}

	task, err := c.tasks.Update(r.Context(), id, callerID, input)
	if err != nil {
		common.SendAppError(w, err)
		return
	}

	common.SendJSON(w, http.StatusOK, task)
}

// Delete godoc
// DELETE /v1/tasks/{id}
func (c *Controller) Delete(w http.ResponseWriter, r *http.Request) {
	callerID := r.Context().Value(common.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid task id", err))
		return
	}

	if err := c.tasks.Delete(r.Context(), id, callerID); err != nil {
		common.SendAppError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
