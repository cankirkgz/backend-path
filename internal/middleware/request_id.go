package middleware

import (
	"context"
	"net/http"
	"strconv"
	"sync/atomic"
	"time"
)

const RequestIDHeader = "X-Request-ID"

type contextKey string

const requestIDContextKey contextKey = "request_id"

var requestCounter uint64

func RequestID(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requestID := r.Header.Get(RequestIDHeader)
		if requestID == "" {
			requestID = generateRequestID()
		}

		ctx := context.WithValue(r.Context(), requestIDContextKey, requestID)
		r = r.WithContext(ctx)

		w.Header().Set(RequestIDHeader, requestID)

		next.ServeHTTP(w, r)
	})
}

func GetRequestID(r *http.Request) string {
	if r == nil {
		return ""
	}

	requestID, ok := r.Context().Value(requestIDContextKey).(string)
	if !ok {
		return ""
	}

	return requestID
}

func generateRequestID() string {
	counter := atomic.AddUint64(&requestCounter, 1)
	return strconv.FormatInt(time.Now().UnixNano(), 10) + "-" + strconv.FormatUint(counter, 10)
}
