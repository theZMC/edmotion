package runtime

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/thezmc/edmotion/internal/config"
)

func TestRunReturnsChallengeLoadError(t *testing.T) {
	r := New(config.Settings{
		ChallengeDir:          filepath.Join(os.TempDir(), "missing-challenge-dir-does-not-exist"),
		HTTPAddr:              ":8080",
		LogLevel:              "info",
		RequestLimitPerMinute: 12,
	})

	err := r.Run()
	if err == nil {
		t.Fatalf("expected run to fail when challenge directory is missing")
	}

	if !strings.Contains(err.Error(), "loading challenges") {
		t.Fatalf("expected loading challenges error, got %q", err.Error())
	}
}

func TestRunReturnsListenErrorForInvalidAddress(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "runtime-challenges-*")
	if err != nil {
		t.Fatalf("MkdirTemp returned error: %v", err)
	}
	defer os.RemoveAll(tmpDir)

	challengeDir := filepath.Join(tmpDir, "sample")
	if err := os.Mkdir(challengeDir, 0o755); err != nil {
		t.Fatalf("Mkdir returned error: %v", err)
	}

	if err := os.WriteFile(filepath.Join(challengeDir, "broken"), []byte("print('broken')\n"), 0o644); err != nil {
		t.Fatalf("writing broken fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(challengeDir, "fixed"), []byte("print('fixed')\n"), 0o644); err != nil {
		t.Fatalf("writing fixed fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(challengeDir, "flag"), []byte("FLAG{sample}\n"), 0o644); err != nil {
		t.Fatalf("writing flag fixture: %v", err)
	}
	if err := os.WriteFile(filepath.Join(challengeDir, "max"), []byte("6\n"), 0o644); err != nil {
		t.Fatalf("writing max fixture: %v", err)
	}

	r := New(config.Settings{
		ChallengeDir:          tmpDir,
		HTTPAddr:              "invalid-address",
		LogLevel:              "info",
		RequestLimitPerMinute: 12,
	})

	err = r.Run()
	if err == nil {
		t.Fatalf("expected run to fail for invalid listen address")
	}

	if !strings.Contains(err.Error(), "server error") {
		t.Fatalf("expected server error, got %q", err.Error())
	}
}
