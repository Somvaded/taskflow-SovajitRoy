package project_controller

import (
	"encoding/json"
	"net/http"
	"strconv"

	"taskflow/delivery/http/common"
	domain_error "taskflow/internal/domain/errors"
	domain_project "taskflow/internal/domain/project"
	"taskflow/utils/validator"

	"github.com/go-chi/chi/v5"
	"github.com/google/uuid"
)

type Controller struct {
	projects domain_project.UseCase
}

func New(projects domain_project.UseCase) *Controller {
	return &Controller{projects: projects}
}

// List godoc
// GET /v1/projects
func (c *Controller) List(w http.ResponseWriter, r *http.Request) {
	callerID := r.Context().Value(common.UserIDKey).(uuid.UUID)

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

	projects, err := c.projects.List(r.Context(), callerID, domain_project.Pagination{Page: page, PageSize: pageSize})
	if err != nil {
		common.SendAppError(w, err)
		return
	}

	if projects == nil {
		projects = []*domain_project.Project{}
	}
	common.SendJSON(w, http.StatusOK, map[string]any{"projects": projects})
}

// Create godoc
// POST /v1/projects
func (c *Controller) Create(w http.ResponseWriter, r *http.Request) {
	callerID := r.Context().Value(common.UserIDKey).(uuid.UUID)

	var req createRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "", err))
		return
	}
	if err := validator.Validate(r.Context(), req); err != nil {
		common.SendAppError(w, err)
		return
	}

	project, err := c.projects.Create(r.Context(), domain_project.CreateInput{
		Name:        req.Name,
		Description: req.Description,
		OwnerID:     callerID,
	})
	if err != nil {
		common.SendAppError(w, err)
		return
	}

	common.SendJSON(w, http.StatusCreated, project)
}

// Get godoc
// GET /v1/projects/{id}
func (c *Controller) Get(w http.ResponseWriter, r *http.Request) {
	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid project id", err))
		return
	}

	project, err := c.projects.Get(r.Context(), id)
	if err != nil {
		common.SendAppError(w, err)
		return
	}

	common.SendJSON(w, http.StatusOK, project)
}

// Update godoc
// PATCH /v1/projects/{id}
func (c *Controller) Update(w http.ResponseWriter, r *http.Request) {
	callerID := r.Context().Value(common.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid project id", err))
		return
	}

	var req updateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "", err))
		return
	}

	project, err := c.projects.Update(r.Context(), id, callerID, domain_project.UpdateInput{
		Name:        req.Name,
		Description: req.Description,
	})
	if err != nil {
		common.SendAppError(w, err)
		return
	}

	common.SendJSON(w, http.StatusOK, project)
}

// Delete godoc
// DELETE /v1/projects/{id}
func (c *Controller) Delete(w http.ResponseWriter, r *http.Request) {
	callerID := r.Context().Value(common.UserIDKey).(uuid.UUID)

	id, err := uuid.Parse(chi.URLParam(r, "id"))
	if err != nil {
		common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "invalid project id", err))
		return
	}

	if err := c.projects.Delete(r.Context(), id, callerID); err != nil {
		common.SendAppError(w, err)
		return
	}

	w.WriteHeader(http.StatusNoContent)
}
