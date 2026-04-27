package admin

import (
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewStateWithProvidedPassword(t *testing.T) {
	state, generated, err := NewState("secret", true)
	if err != nil {
		t.Fatalf("NewState returned error: %v", err)
	}

	if generated != "" {
		t.Fatalf("expected empty generated password, got %q", generated)
	}

	if !state.ServeFixedFiles() {
		t.Fatalf("expected fixed files to be enabled")
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", "secret")
	if !state.IsAuthorized(req) {
		t.Fatalf("expected request to be authorized")
	}
}

func TestNewStateGeneratesPasswordWhenMissing(t *testing.T) {
	state, generated, err := NewState("", false)
	if err != nil {
		t.Fatalf("NewState returned error: %v", err)
	}

	if len(generated) != 32 {
		t.Fatalf("expected 32-char generated password, got %d", len(generated))
	}

	if strings.TrimSpace(generated) == "" {
		t.Fatalf("expected non-empty generated password")
	}

	req := httptest.NewRequest("GET", "/", nil)
	req.Header.Set("Authorization", generated)
	if !state.IsAuthorized(req) {
		t.Fatalf("expected generated password to authorize request")
	}
}

func TestToggleFixedFiles(t *testing.T) {
	state, _, err := NewState("secret", false)
	if err != nil {
		t.Fatalf("NewState returned error: %v", err)
	}

	if state.ToggleFixedFiles() != true {
		t.Fatalf("expected first toggle to enable fixed files")
	}

	if state.ServeFixedFiles() != true {
		t.Fatalf("expected fixed files to remain enabled")
	}

	if state.ToggleFixedFiles() != false {
		t.Fatalf("expected second toggle to disable fixed files")
	}
}

func TestSetPassword(t *testing.T) {
	state, _, err := NewState("initial", false)
	if err != nil {
		t.Fatalf("NewState returned error: %v", err)
	}

	state.SetPassword("updated")

	oldReq := httptest.NewRequest("GET", "/", nil)
	oldReq.Header.Set("Authorization", "initial")
	if state.IsAuthorized(oldReq) {
		t.Fatalf("expected old password to be rejected")
	}

	newReq := httptest.NewRequest("GET", "/", nil)
	newReq.Header.Set("Authorization", "updated")
	if !state.IsAuthorized(newReq) {
		t.Fatalf("expected new password to be accepted")
	}
}
