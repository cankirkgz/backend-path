package middleware

import (
	"log/slog"
	"net/http"
	"time"
)

type statusRecorder struct {
	http.ResponseWriter
	statusCode int
}

func newStatusRecorder(w http.ResponseWriter) *statusRecorder {
	return &statusRecorder{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (r *statusRecorder) WriteHeader(statusCode int) {
	r.statusCode = statusCode
	r.ResponseWriter.WriteHeader(statusCode)
}

func Logging(logg *slog.Logger) Middleware {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			start := time.Now()

			recorder := newStatusRecorder(w)

			next.ServeHTTP(recorder, r)

			duration := time.Since(start)
			requestID := GetRequestID(r)
			userID := GetAuthenticatedUserID(r)
			userRole := GetAuthenticatedUserRole(r)

			attrs := []any{
				"method", r.Method,
				"path", r.URL.Path,
				"status", recorder.statusCode,
				"duration_ms", duration.Milliseconds(),
				"request_id", requestID,
				"remote_addr", r.RemoteAddr,
				"user_agent", r.UserAgent(),
			}

			if userID != 0 {
				attrs = append(attrs, "user_id", userID)
			}

			if userRole != "" {
				attrs = append(attrs, "user_role", userRole)
			}

			if recorder.statusCode >= 500 {
				logg.Error("http request completed", attrs...)
				return
			}

			if recorder.statusCode >= 400 {
				logg.Warn("http request completed", attrs...)
				return
			}

			logg.Info("http request completed", attrs...)
		})
	}
}
