package auth_controller

import (
	"encoding/json"
	"net/http"

	"taskflow/delivery/http/common"
	domain_error "taskflow/internal/domain/errors"
	domain_user "taskflow/internal/domain/user"
	"taskflow/utils/validator"
)

type Controller struct {
	auth domain_user.UseCase
}

func New(auth domain_user.UseCase) *Controller {
	return &Controller{auth: auth}
}

// Register godoc
// POST /auth/register
func (c *Controller) Register(w http.ResponseWriter, r *http.Request) {
	var req registerRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "", err))
		return
	}
	if err := validator.Validate(r.Context(), req); err != nil {
		common.SendAppError(w, err)
		return
	}

	result, err := c.auth.Register(r.Context(), domain_user.RegisterInput{
		Name:     req.Name,
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		common.SendAppError(w, err)
		return
	}

	common.SendJSON(w, http.StatusCreated, map[string]any{
		"token": result.Token,
		"user":  result.User,
	})
}

// Login godoc
// POST /auth/login
func (c *Controller) Login(w http.ResponseWriter, r *http.Request) {
	var req loginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		common.SendAppError(w, domain_error.Raise(domain_error.CODE_INVALID_PAYLOAD, "", err))
		return
	}
	if err := validator.Validate(r.Context(), req); err != nil {
		common.SendAppError(w, err)
		return
	}

	result, err := c.auth.Login(r.Context(), domain_user.LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		common.SendAppError(w, err)
		return
	}

	common.SendJSON(w, http.StatusOK, map[string]any{
		"token": result.Token,
		"user":  result.User,
	})
}
