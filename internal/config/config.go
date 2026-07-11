package config

import "os"

type Config struct {
	Env         string
	HTTPAddr    string
	DatabaseURL string
	StaticDir   string
}

func Load() (Config, error) {
	cfg := Config{
		Env:         getenv("APP_ENV", "local"),
		HTTPAddr:    httpAddr(),
		DatabaseURL: os.Getenv("DATABASE_URL"),
		StaticDir:   os.Getenv("STATIC_DIR"),
	}

	return cfg, nil
}

func httpAddr() string {
	if port := os.Getenv("PORT"); port != "" {
		return ":" + port
	}
	if value := os.Getenv("HTTP_ADDR"); value != "" {
		return value
	}
	return ":8080"
}

func getenv(key string, fallback string) string {
	value := os.Getenv(key)
	if value == "" {
		return fallback
	}
	return value
}
