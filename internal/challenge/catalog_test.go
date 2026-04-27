package challenge

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
)

func TestCatalogReloadIfChanged(t *testing.T) {
	tmpDir := t.TempDir()
	writeChallengeDir(t, tmpDir, "alpha", "6", "FLAG{alpha}")

	catalog, err := NewCatalog(tmpDir)
	if err != nil {
		t.Fatalf("NewCatalog returned error: %v", err)
	}

	changed, err := catalog.ReloadIfChanged()
	if err != nil {
		t.Fatalf("ReloadIfChanged returned error: %v", err)
	}
	if changed {
		t.Fatalf("expected no changes immediately after creation")
	}

	writeChallengeDir(t, tmpDir, "beta", "7", "FLAG{beta}")

	changed, err = catalog.ReloadIfChanged()
	if err != nil {
		t.Fatalf("ReloadIfChanged returned error: %v", err)
	}
	if !changed {
		t.Fatalf("expected reload change after adding a challenge")
	}

	if catalog.Len() != 2 {
		t.Fatalf("expected catalog length 2, got %d", catalog.Len())
	}
}

func TestCatalogRoutesUseLatestChallengeInfo(t *testing.T) {
	tmpDir := t.TempDir()
	writeChallengeDir(t, tmpDir, "alpha", "6", "FLAG{alpha}")

	catalog, err := NewCatalog(tmpDir)
	if err != nil {
		t.Fatalf("NewCatalog returned error: %v", err)
	}

	r := chi.NewRouter()
	catalog.RegisterRoutes(r)

	initialReq := httptest.NewRequest(http.MethodGet, "/alpha", nil)
	initialRR := httptest.NewRecorder()
	r.ServeHTTP(initialRR, initialReq)
	if initialRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", initialRR.Code)
	}
	if !strings.Contains(initialRR.Body.String(), "Max Characters: 6") {
		t.Fatalf("expected initial max characters in response")
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "alpha", "max"), []byte("12\n"), 0o644); err != nil {
		t.Fatalf("updating max fixture: %v", err)
	}

	changed, err := catalog.ReloadIfChanged()
	if err != nil {
		t.Fatalf("ReloadIfChanged returned error: %v", err)
	}
	if !changed {
		t.Fatalf("expected reload change after max update")
	}

	reloadedReq := httptest.NewRequest(http.MethodGet, "/alpha", nil)
	reloadedRR := httptest.NewRecorder()
	r.ServeHTTP(reloadedRR, reloadedReq)
	if reloadedRR.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", reloadedRR.Code)
	}
	if !strings.Contains(reloadedRR.Body.String(), "Max Characters: 12") {
		t.Fatalf("expected reloaded max characters in response")
	}
}

func TestCatalogReloadStatsBySource(t *testing.T) {
	tmpDir := t.TempDir()
	writeChallengeDir(t, tmpDir, "alpha", "6", "FLAG{alpha}")

	catalog, err := NewCatalog(tmpDir)
	if err != nil {
		t.Fatalf("NewCatalog returned error: %v", err)
	}

	catalog.reloadAndLog("poll")
	stats := catalog.Stats()
	if stats.PollAttempts != 1 {
		t.Fatalf("expected 1 poll attempt, got %d", stats.PollAttempts)
	}
	if stats.WatcherAttempts != 0 {
		t.Fatalf("expected 0 watcher attempts, got %d", stats.WatcherAttempts)
	}
	if stats.ChangedReloads != 0 {
		t.Fatalf("expected 0 changed reloads, got %d", stats.ChangedReloads)
	}

	if err := os.WriteFile(filepath.Join(tmpDir, "alpha", "max"), []byte("12\n"), 0o644); err != nil {
		t.Fatalf("updating max fixture: %v", err)
	}

	catalog.reloadAndLog("watcher")
	stats = catalog.Stats()
	if stats.WatcherAttempts != 1 {
		t.Fatalf("expected 1 watcher attempt, got %d", stats.WatcherAttempts)
	}
	if stats.PollAttempts != 1 {
		t.Fatalf("expected poll attempts to remain 1, got %d", stats.PollAttempts)
	}
	if stats.ChangedReloads != 1 {
		t.Fatalf("expected 1 changed reload, got %d", stats.ChangedReloads)
	}
}

func writeChallengeDir(t *testing.T, root string, id string, max string, flag string) {
	t.Helper()

	dir := filepath.Join(root, id)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		t.Fatalf("MkdirAll returned error: %v", err)
	}

	if err := os.WriteFile(filepath.Join(dir, "broken"), []byte("print('broken')\n"), 0o644); err != nil {
		t.Fatalf("writing broken fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "fixed"), []byte("print('fixed')\n"), 0o644); err != nil {
		t.Fatalf("writing fixed fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "max"), []byte(max+"\n"), 0o644); err != nil {
		t.Fatalf("writing max fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(dir, "flag"), []byte(flag+"\n"), 0o644); err != nil {
		t.Fatalf("writing flag fixture: %v", err)
	}
}
