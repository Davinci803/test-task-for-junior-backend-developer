package docs

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestDocsHandler(t *testing.T) {
	h := NewHandler()

	rec := httptest.NewRecorder()
	h.ServeSpec(rec, httptest.NewRequest(http.MethodGet, "/swagger/openapi.json", nil))
	if rec.Code != http.StatusOK {
		t.Fatalf("unexpected status: %d", rec.Code)
	}
	if !strings.Contains(rec.Body.String(), "\"openapi\"") {
		t.Fatalf("spec body does not look like openapi")
	}

	rec = httptest.NewRecorder()
	h.ServeUI(rec, httptest.NewRequest(http.MethodGet, "/swagger/", nil))
	if rec.Code != http.StatusOK || !strings.Contains(rec.Body.String(), "SwaggerUIBundle") {
		t.Fatalf("unexpected swagger ui response")
	}

	rec = httptest.NewRecorder()
	h.RedirectToUI(rec, httptest.NewRequest(http.MethodGet, "/swagger", nil))
	if rec.Code != http.StatusMovedPermanently {
		t.Fatalf("unexpected redirect status: %d", rec.Code)
	}
}
