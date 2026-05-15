package app

import "testing"

func TestParseConfigDefaults(t *testing.T) {
	cfg, err := ParseConfig(nil)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "127.0.0.1" {
		t.Fatalf("expected default host, got %s", cfg.Host)
	}
	if cfg.Port != 8421 {
		t.Fatalf("expected default port, got %d", cfg.Port)
	}
}

func TestParseConfigPublicOverridesHost(t *testing.T) {
	cfg, err := ParseConfig([]string{"--host", "127.0.0.1", "--public"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if cfg.Host != "0.0.0.0" {
		t.Fatalf("expected 0.0.0.0, got %s", cfg.Host)
	}
}

func TestParseConfigRejectsInvalidPort(t *testing.T) {
	_, err := ParseConfig([]string{"--port", "70000"})
	if err == nil {
		t.Fatal("expected error for invalid port")
	}
}
