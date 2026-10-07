package platform

import (
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"
)

type Config struct {
	Address, DatabaseURL, AllowedOrigins, LogLevel string
	AutoMigrate                                    bool
	Environment                                    string
	EnableDocs                                     bool
	BFFSecret                                      string
	MetricsToken                                   string
}

func NewConfig() (Config, error) {
	c := Config{Environment: env("APP_ENV", "development"), Address: env("HTTP_ADDRESS", "127.0.0.1:8081"), DatabaseURL: os.Getenv("DATABASE_URL"), AllowedOrigins: env("CORS_ORIGINS", "http://127.0.0.1:5173,http://localhost:5173"), LogLevel: env("LOG_LEVEL", "info")}
	if c.DatabaseURL == "" {
		return c, fmt.Errorf("DATABASE_URL is required")
	}
	c.BFFSecret = os.Getenv("BFF_SHARED_SECRET")
	c.MetricsToken = os.Getenv("METRICS_TOKEN")
	var err error
	c.AutoMigrate, err = strconv.ParseBool(env("AUTO_MIGRATE", "false"))
	if err != nil {
		return c, fmt.Errorf("AUTO_MIGRATE must be a boolean")
	}
	c.EnableDocs, err = strconv.ParseBool(env("ENABLE_API_DOCS", strconv.FormatBool(c.Environment != "production")))
	if err != nil {
		return c, fmt.Errorf("ENABLE_API_DOCS must be boolean")
	}
	if c.Environment != "development" && c.Environment != "test" && c.Environment != "production" {
		return c, fmt.Errorf("invalid APP_ENV")
	}
	if c.Environment == "production" {
		if len(c.MetricsToken) < 32 {
			return c, fmt.Errorf("production requires METRICS_TOKEN of at least 32 bytes")
		}
		if len(c.BFFSecret) < 32 {
			return c, fmt.Errorf("production requires BFF_SHARED_SECRET of at least 32 bytes")
		}
		if c.AutoMigrate {
			return c, fmt.Errorf("run migrations separately; AUTO_MIGRATE must be false in production")
		}
		db, err := url.Parse(c.DatabaseURL)
		if err != nil || db.User == nil || db.Query().Get("sslmode") != "verify-full" {
			return c, fmt.Errorf("production database requires sslmode=verify-full")
		}
		password, _ := db.User.Password()
		if password == "" || password == "forma_local" {
			return c, fmt.Errorf("production database requires a non-default credential")
		}
		if os.Getenv("CORS_ORIGINS") == "" {
			return c, fmt.Errorf("CORS_ORIGINS required in production")
		}
	}
	for _, origin := range strings.Split(c.AllowedOrigins, ",") {
		u, err := url.Parse(origin)
		if err != nil || u.Host == "" || u.User != nil || u.Path != "" || u.RawQuery != "" || u.Fragment != "" || strings.Contains(origin, "*") || (u.Scheme != "https" && (u.Scheme != "http" || c.Environment == "production")) {
			return c, fmt.Errorf("CORS_ORIGINS must contain explicit valid origins")
		}
	}
	return c, nil
}
func env(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
