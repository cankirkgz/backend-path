package middleware

import (
	"context"
	"net/http"
	"strings"

	"backend-path/internal/domain"
	"backend-path/internal/service"
)

type authContextKey string

const (
	userIDContextKey   authContextKey = "user_id"
	userRoleContextKey authContextKey = "user_role"
)

func Authentication(tokenService *service.TokenService) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			authHeader := strings.TrimSpace(r.Header.Get("Authorization"))
			if authHeader == "" {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"missing authorization header"}`))
				return
			}

			const bearerPrefix = "Bearer "
			if !strings.HasPrefix(authHeader, bearerPrefix) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"invalid authorization header format"}`))
				return
			}

			token := strings.TrimSpace(strings.TrimPrefix(authHeader, bearerPrefix))
			payload, err := tokenService.Parse(token)
			if err != nil {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				_, _ = w.Write([]byte(`{"error":"invalid or expired token"}`))
				return
			}

			ctx := context.WithValue(r.Context(), userIDContextKey, payload.UserID)
			ctx = context.WithValue(ctx, userRoleContextKey, payload.Role)

			next.ServeHTTP(w, r.WithContext(ctx))
		})
	}
}

func GetAuthenticatedUserID(r *http.Request) int64 {
	if r == nil {
		return 0
	}

	userID, ok := r.Context().Value(userIDContextKey).(int64)
	if !ok {
		return 0
	}

	return userID
}

func GetAuthenticatedUserRole(r *http.Request) domain.UserRole {
	if r == nil {
		return ""
	}

	role, ok := r.Context().Value(userRoleContextKey).(domain.UserRole)
	if !ok {
		return ""
	}

	return role
}
