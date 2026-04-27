package config

import "testing"

func TestLoadDefaults(t *testing.T) {
	t.Setenv("GIVE_FIXED_FILES", "")
	t.Setenv("CHALLENGE_DIR", "")
	t.Setenv("ADMIN_PASSWORD", "")
	t.Setenv("HTTP_ADDR", "")
	t.Setenv("LOG_LEVEL", "")
	t.Setenv("REQUEST_LIMIT_PER_MINUTE", "")

	settings := Load()

	if settings.GiveFixedFiles {
		t.Fatalf("expected GiveFixedFiles to default to false")
	}
	if settings.ChallengeDir != "challenges" {
		t.Fatalf("expected default ChallengeDir to be challenges, got %q", settings.ChallengeDir)
	}
	if settings.AdminPassword != "" {
		t.Fatalf("expected empty AdminPassword by default")
	}
	if settings.HTTPAddr != ":8080" {
		t.Fatalf("expected default HTTPAddr to be :8080, got %q", settings.HTTPAddr)
	}
	if settings.LogLevel != "info" {
		t.Fatalf("expected default LogLevel to be info, got %q", settings.LogLevel)
	}
	if settings.RequestLimitPerMinute != defaultRequestLimitPerMinute {
		t.Fatalf("expected default RequestLimitPerMinute to be %d, got %d", defaultRequestLimitPerMinute, settings.RequestLimitPerMinute)
	}
}

func TestLoadFromEnvironment(t *testing.T) {
	t.Setenv("GIVE_FIXED_FILES", "true")
	t.Setenv("CHALLENGE_DIR", "custom-challenges")
	t.Setenv("ADMIN_PASSWORD", "admin-secret")
	t.Setenv("HTTP_ADDR", "127.0.0.1:9090")
	t.Setenv("LOG_LEVEL", "debug")
	t.Setenv("REQUEST_LIMIT_PER_MINUTE", "33")

	settings := Load()

	if !settings.GiveFixedFiles {
		t.Fatalf("expected GiveFixedFiles to be true")
	}
	if settings.ChallengeDir != "custom-challenges" {
		t.Fatalf("expected ChallengeDir to come from env, got %q", settings.ChallengeDir)
	}
	if settings.AdminPassword != "admin-secret" {
		t.Fatalf("expected AdminPassword to come from env, got %q", settings.AdminPassword)
	}
	if settings.HTTPAddr != "127.0.0.1:9090" {
		t.Fatalf("expected HTTPAddr to come from env, got %q", settings.HTTPAddr)
	}
	if settings.LogLevel != "debug" {
		t.Fatalf("expected LogLevel to come from env, got %q", settings.LogLevel)
	}
	if settings.RequestLimitPerMinute != 33 {
		t.Fatalf("expected RequestLimitPerMinute to come from env, got %d", settings.RequestLimitPerMinute)
	}
}

func TestLoadInvalidRequestLimitFallsBackToDefault(t *testing.T) {
	t.Setenv("REQUEST_LIMIT_PER_MINUTE", "not-a-number")

	settings := Load()

	if settings.RequestLimitPerMinute != defaultRequestLimitPerMinute {
		t.Fatalf("expected invalid request limit to fall back to %d, got %d", defaultRequestLimitPerMinute, settings.RequestLimitPerMinute)
	}
}
