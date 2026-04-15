package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestPerformanceAddsResponseTimeHeader(t *testing.T) {
	handler := Performance(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	headerValue := rec.Header().Get(ResponseTimeHeader)
	if headerValue == "" {
		t.Fatalf("expected %s header to be set", ResponseTimeHeader)
	}
}

func TestPerformanceWorksWhenHandlerOnlyWritesBody(t *testing.T) {
	handler := Performance(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte("ok"))
	}))

	req := httptest.NewRequest(http.MethodGet, "/health", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected status %d, got %d", http.StatusOK, rec.Code)
	}

	headerValue := rec.Header().Get(ResponseTimeHeader)
	if headerValue == "" {
		t.Fatalf("expected %s header to be set", ResponseTimeHeader)
	}
}
