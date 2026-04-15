package middleware

import (
	"net/http"
	"strconv"
	"time"
)

const ResponseTimeHeader = "X-Response-Time-Ms"

type performanceRecorder struct {
	http.ResponseWriter
	start time.Time
}

func newPerformanceRecorder(w http.ResponseWriter) *performanceRecorder {
	return &performanceRecorder{
		ResponseWriter: w,
		start:          time.Now(),
	}
}

func (r *performanceRecorder) WriteHeader(statusCode int) {
	duration := time.Since(r.start)
	r.Header().Set(ResponseTimeHeader, strconv.FormatInt(duration.Milliseconds(), 10))
	r.ResponseWriter.WriteHeader(statusCode)
}

func (r *performanceRecorder) Write(data []byte) (int, error) {
	if r.Header().Get(ResponseTimeHeader) == "" {
		duration := time.Since(r.start)
		r.Header().Set(ResponseTimeHeader, strconv.FormatInt(duration.Milliseconds(), 10))
	}

	return r.ResponseWriter.Write(data)
}

func Performance(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		recorder := newPerformanceRecorder(w)
		next.ServeHTTP(recorder, r)
	})
}
