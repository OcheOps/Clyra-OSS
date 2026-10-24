package server

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestRouteBodySizeAllowsCloudSaveAboveDefaultLimit(t *testing.T) {
	reader := func(w http.ResponseWriter, r *http.Request) {
		_, err := io.ReadAll(r.Body)
		var tooLarge *http.MaxBytesError
		if errors.As(err, &tooLarge) {
			http.Error(w, "too large", http.StatusRequestEntityTooLarge)
			return
		}
		if err != nil {
			http.Error(w, err.Error(), http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusNoContent)
	}
	handler := routeBodySize(10)(http.HandlerFunc(reader))
	body := strings.Repeat("x", 11)

	cloudReq := httptest.NewRequest(http.MethodPut, "/api/v1/sync/projects/project-1", strings.NewReader(body))
	cloudRes := httptest.NewRecorder()
	handler.ServeHTTP(cloudRes, cloudReq)
	if cloudRes.Code != http.StatusNoContent {
		t.Fatalf("cloud save status=%d, want 204", cloudRes.Code)
	}

	regularReq := httptest.NewRequest(http.MethodPost, "/api/v1/feedback", strings.NewReader(body))
	regularRes := httptest.NewRecorder()
	handler.ServeHTTP(regularRes, regularReq)
	if regularRes.Code != http.StatusRequestEntityTooLarge {
		t.Fatalf("regular route status=%d, want 413", regularRes.Code)
	}
}
