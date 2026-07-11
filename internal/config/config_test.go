package config

import "testing"

func TestLoadUsesPortForHTTPAddr(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://calendar:calendar@localhost:5432/calendar")
	t.Setenv("PORT", "10000")
	t.Setenv("HTTP_ADDR", ":8080")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.HTTPAddr != ":10000" {
		t.Fatalf("expected PORT to define HTTPAddr, got %q", cfg.HTTPAddr)
	}
}

func TestLoadFallsBackToHTTPAddr(t *testing.T) {
	t.Setenv("DATABASE_URL", "postgres://calendar:calendar@localhost:5432/calendar")
	t.Setenv("HTTP_ADDR", ":9090")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.HTTPAddr != ":9090" {
		t.Fatalf("expected HTTP_ADDR fallback, got %q", cfg.HTTPAddr)
	}
}

func TestLoadAllowsEmptyDatabaseURL(t *testing.T) {
	t.Setenv("DATABASE_URL", "")

	cfg, err := Load()
	if err != nil {
		t.Fatal(err)
	}

	if cfg.DatabaseURL != "" {
		t.Fatalf("expected empty DatabaseURL, got %q", cfg.DatabaseURL)
	}
}
