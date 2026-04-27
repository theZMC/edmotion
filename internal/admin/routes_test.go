package admin

import (
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestRegisterRoutesToggleFixedFilesAuthorization(t *testing.T) {
	state, _, err := NewState("secret", false)
	if err != nil {
		t.Fatalf("NewState returned error: %v", err)
	}

	r := chi.NewRouter()
	RegisterRoutes(r, state)

	req := httptest.NewRequest(http.MethodPut, "/admin/toggle-fixed-files", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d", rr.Code)
	}
}

func TestRegisterRoutesToggleFixedFilesSuccess(t *testing.T) {
	state, _, err := NewState("secret", false)
	if err != nil {
		t.Fatalf("NewState returned error: %v", err)
	}

	r := chi.NewRouter()
	RegisterRoutes(r, state)

	req := httptest.NewRequest(http.MethodPut, "/admin/toggle-fixed-files", nil)
	req.Header.Set("Authorization", "secret")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	if !strings.Contains(rr.Body.String(), "giveFixedFiles set to true") {
		t.Fatalf("expected toggle response body, got %q", rr.Body.String())
	}

	if !state.ServeFixedFiles() {
		t.Fatalf("expected state to toggle fixed files on")
	}
}

func TestRegisterRoutesSetPasswordValidation(t *testing.T) {
	state, _, err := NewState("secret", false)
	if err != nil {
		t.Fatalf("NewState returned error: %v", err)
	}

	r := chi.NewRouter()
	RegisterRoutes(r, state)

	req := httptest.NewRequest(http.MethodPut, "/admin/set-password", strings.NewReader(url.Values{"password": {""}}.Encode()))
	req.Header.Set("Authorization", "secret")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d", rr.Code)
	}
}

func TestRegisterRoutesSetPasswordSuccess(t *testing.T) {
	state, _, err := NewState("secret", false)
	if err != nil {
		t.Fatalf("NewState returned error: %v", err)
	}

	r := chi.NewRouter()
	RegisterRoutes(r, state)

	setReq := httptest.NewRequest(http.MethodPut, "/admin/set-password", strings.NewReader(url.Values{"password": {"next"}}.Encode()))
	setReq.Header.Set("Authorization", "secret")
	setReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	setRR := httptest.NewRecorder()
	r.ServeHTTP(setRR, setReq)

	if setRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", setRR.Code)
	}

	oldAuthReq := httptest.NewRequest(http.MethodPut, "/admin/toggle-fixed-files", nil)
	oldAuthReq.Header.Set("Authorization", "secret")
	oldAuthRR := httptest.NewRecorder()
	r.ServeHTTP(oldAuthRR, oldAuthReq)
	if oldAuthRR.Code != http.StatusUnauthorized {
		t.Fatalf("expected old password to fail with 401, got %d", oldAuthRR.Code)
	}

	newAuthReq := httptest.NewRequest(http.MethodPut, "/admin/toggle-fixed-files", nil)
	newAuthReq.Header.Set("Authorization", "next")
	newAuthRR := httptest.NewRecorder()
	r.ServeHTTP(newAuthRR, newAuthReq)
	if newAuthRR.Code != http.StatusOK {
		t.Fatalf("expected new password to succeed with 200, got %d", newAuthRR.Code)
	}
}
