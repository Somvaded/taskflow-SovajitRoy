package middleware

import (
	"context"
	"net/http"
	"runtime/debug"
	"time"

	"taskflow/delivery/http/common"
	"taskflow/utils/logger"

	"github.com/google/uuid"
)

// RequestID generates and injects a unique request ID.
func RequestID() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			requestID := r.Header.Get("X-Request-ID")
			if requestID == "" {
				requestID = uuid.New().String()
			}
			w.Header().Set("X-Request-ID", requestID)
			ctx := context.WithValue(r.Context(), common.RequestIDKey, requestID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

// Logger injects a request-scoped logger into context and logs the request/response.
func Logger() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()
			requestID, _ := r.Context().Value(common.RequestIDKey).(string)

			// Enrich logger with request fields and store in context.
			log := logger.FromContext(r.Context()).With(
				"request_id", requestID,
				"method", r.Method,
				"path", r.URL.Path,
				"remote_addr", r.RemoteAddr,
			)
			ctx := logger.NewContext(r.Context(), log)

			log.InfoContext(ctx, "request")

			wrw := &wrappedWriter{ResponseWriter: w, statusCode: http.StatusOK}
			next.ServeHTTP(wrw, r.WithContext(ctx))

			log.InfoContext(ctx, "response",
				"status", wrw.statusCode,
				"duration", time.Since(start),
			)
		})
	}
}

// Recovery catches panics and returns 500.
func Recovery() func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				if err := recover(); err != nil {
					logger.FromContext(r.Context()).Error("panic recovered",
						"error", err,
						"stack", string(debug.Stack()),
					)
					w.Header().Set("Content-Type", "application/json")
					w.WriteHeader(http.StatusInternalServerError)
					w.Write([]byte(`{"code":"FAILURE","message":"Internal server error"}`))
				}
			}()
			next.ServeHTTP(w, r)
		})
	}
}

type wrappedWriter struct {
	http.ResponseWriter
	statusCode int
}

func (w *wrappedWriter) WriteHeader(code int) {
	w.statusCode = code
	w.ResponseWriter.WriteHeader(code)
}
