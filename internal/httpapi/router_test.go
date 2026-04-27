package httpapi

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thezmc/edmotion/internal/admin"
	"github.com/thezmc/edmotion/internal/challenge"
)

func TestNewRouterRequiresAdminState(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "httpapi-catalog-*")
	if err != nil {
		t.Fatalf("MkdirTemp returned error: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	catalog, err := challenge.NewCatalog(tmpDir)
	if err != nil {
		t.Fatalf("NewCatalog returned error: %v", err)
	}

	_, err = NewRouter(Params{Catalog: catalog})
	if err == nil {
		t.Fatalf("expected error when admin state is missing")
	}
}

func TestNewRouterRequiresChallengeCatalog(t *testing.T) {
	state, _, err := admin.NewState("secret", false)
	if err != nil {
		t.Fatalf("NewState returned error: %v", err)
	}

	_, err = NewRouter(Params{Admin: state})
	if err == nil {
		t.Fatalf("expected error when challenge catalog is missing")
	}
}

func TestNewRouterRegistersAdminEndpoints(t *testing.T) {
	state, _, err := admin.NewState("secret", false)
	if err != nil {
		t.Fatalf("NewState returned error: %v", err)
	}

	tmpDir, err := os.MkdirTemp("", "httpapi-catalog-*")
	if err != nil {
		t.Fatalf("MkdirTemp returned error: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	catalog, err := challenge.NewCatalog(tmpDir)
	if err != nil {
		t.Fatalf("NewCatalog returned error: %v", err)
	}

	r, err := NewRouter(Params{Catalog: catalog, Admin: state, RequestLimitPerMinute: 1000})
	if err != nil {
		t.Fatalf("NewRouter returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodPut, "/admin/toggle-fixed-files", nil)
	req.Header.Set("Authorization", "secret")
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
}

func TestNewRouterRegistersChallengeRoutesAndInjectsFixedContext(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "httpapi-challenge-*")
	if err != nil {
		t.Fatalf("MkdirTemp returned error: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	challengeDir := filepath.Join(tmpDir, "signal-intercept")
	if err := os.Mkdir(challengeDir, 0o755); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}

	if err := os.WriteFile(filepath.Join(challengeDir, "broken"), []byte("print('broken')\n"), 0o644); err != nil {
		t.Fatalf("writing broken fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(challengeDir, "fixed"), []byte("print('fixed')\n"), 0o644); err != nil {
		t.Fatalf("writing fixed fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(challengeDir, "max"), []byte("20\n"), 0o644); err != nil {
		t.Fatalf("writing max fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(challengeDir, "flag"), []byte("FLAG{test}\n"), 0o644); err != nil {
		t.Fatalf("writing flag fixture: %v", err)
	}

	state, _, err := admin.NewState("secret", true)
	if err != nil {
		t.Fatalf("NewState returned error: %v", err)
	}

	catalog, err := challenge.NewCatalog(tmpDir)
	if err != nil {
		t.Fatalf("NewCatalog returned error: %v", err)
	}

	r, err := NewRouter(Params{Catalog: catalog, Admin: state, RequestLimitPerMinute: 1000})
	if err != nil {
		t.Fatalf("NewRouter returned error: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/signal-intercept", nil)
	rr := httptest.NewRecorder()

	r.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}

	body := rr.Body.String()
	if !strings.Contains(body, "-----BEGIN BROKEN APPLICATION-----") {
		t.Fatalf("expected broken application section in response")
	}
	if !strings.Contains(body, "-----BEGIN FIXED APPLICATION-----") {
		t.Fatalf("expected fixed application section in response")
	}
}
