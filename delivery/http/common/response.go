package common

import (
	"encoding/json"
	"net/http"

	domain_error "taskflow/internal/domain/errors"
)

type response struct {
	Data    any    `json:"data,omitempty"`
	Code    string `json:"code"`
	Message string `json:"message,omitempty"`
	Errors  any    `json:"errors,omitempty"`
}

const (
	codeSuccess = "SUCCESS"
	codeFailure = "FAILURE"
)

func SendJSON(w http.ResponseWriter, statusCode int, data any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response{Data: data, Code: codeSuccess})
}

func SendError(w http.ResponseWriter, statusCode int, message string, details any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(response{Code: codeFailure, Message: message, Errors: details})
}

// SendAppError converts an AppError to its HTTP status and writes the response.
func SendAppError(w http.ResponseWriter, err error) {
	httpErr := domain_error.GetHTTPError(err)
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(httpErr.Status)
	json.NewEncoder(w).Encode(httpErr)
}
