package middleware

import (
	"context"
	"log/slog"
	"net/http"
	"strings"

	"taskflow/delivery/http/common"
	"taskflow/utils"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
)

// Authenticate validates the JWT Bearer token and injects user_id into context.
func Authenticate(config *utils.Config) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := r.Header.Get("Authorization")
			requestID, _ := r.Context().Value(common.RequestIDKey).(string)

			if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
				slog.WarnContext(r.Context(), "missing or malformed authorization header",
					"path", r.URL.Path,
					"request_id", requestID,
				)
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"code":"FAILURE","message":"Unauthorized"}`))
				return
			}

			tokenStr := strings.TrimPrefix(authHeader, "Bearer ")
			token, err := jwt.Parse(tokenStr, func(t *jwt.Token) (any, error) {
				if _, ok := t.Method.(*jwt.SigningMethodHMAC); !ok {
					return nil, jwt.ErrSignatureInvalid
				}
				return []byte(config.JWTSecret), nil
			})
			if err != nil || !token.Valid {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"code":"FAILURE","message":"Unauthorized"}`))
				return
			}

			claims, ok := token.Claims.(jwt.MapClaims)
			if !ok {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"code":"FAILURE","message":"Unauthorized"}`))
				return
			}

			userIDStr, _ := claims["user_id"].(string)
			userID, err := uuid.Parse(userIDStr)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				w.Write([]byte(`{"code":"FAILURE","message":"Unauthorized"}`))
				return
			}

			ctx := context.WithValue(r.Context(), common.UserIDKey, userID)
			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}
