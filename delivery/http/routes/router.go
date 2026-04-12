package routes

import (
	"net/http"
	"time"

	"taskflow/delivery/http/common"
	"taskflow/delivery/http/middleware"
	auth_controller "taskflow/delivery/http/controller/auth"
	project_controller "taskflow/delivery/http/controller/project"
	task_controller "taskflow/delivery/http/controller/task"
	"taskflow/internal/usecase"
	"taskflow/utils"

	"github.com/go-chi/chi/v5"
)

func NewRouter(config *utils.Config, uc *usecase.UseCases) http.Handler {
	r := chi.NewRouter()

	// Global middleware
	r.Use(middleware.RequestID())
	r.Use(middleware.Logger())
	r.Use(middleware.Recovery())

	// Public routes
	r.Get("/health", func(w http.ResponseWriter, r *http.Request) {
		common.SendJSON(w, http.StatusOK, map[string]any{
			"status":    "ok",
			"service":   "taskflow",
			"timestamp": time.Now().Unix(),
		})
	})

	// Auth
	authCtrl := auth_controller.New(uc.Auth)
	r.Post("/auth/register", authCtrl.Register)
	r.Post("/auth/login", authCtrl.Login)

	// Authenticated v1 routes
	r.Group(func(r chi.Router) {
		r.Use(middleware.Authenticate(config))

		projectCtrl := project_controller.New(uc.Projects)
		r.Get("/v1/projects", projectCtrl.List)
		r.Post("/v1/projects", projectCtrl.Create)
		r.Get("/v1/projects/{id}", projectCtrl.Get)
		r.Patch("/v1/projects/{id}", projectCtrl.Update)
		r.Delete("/v1/projects/{id}", projectCtrl.Delete)

		taskCtrl := task_controller.New(uc.Tasks)
		r.Get("/v1/projects/{projectID}/tasks", taskCtrl.List)
		r.Post("/v1/projects/{projectID}/tasks", taskCtrl.Create)
		r.Patch("/v1/tasks/{id}", taskCtrl.Update)
		r.Delete("/v1/tasks/{id}", taskCtrl.Delete)
	})

	return r
}
