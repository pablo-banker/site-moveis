package platform

import (
	"strings"
	"testing"
)

func TestProductionConfigurationRejectsUnsafeDefaults(t *testing.T) {
	t.Setenv("APP_ENV", "production")
	t.Setenv("BFF_SHARED_SECRET", strings.Repeat("a", 32))
	t.Setenv("METRICS_TOKEN", strings.Repeat("b", 32))
	t.Setenv("DATABASE_URL", "postgres://runtime:unique-secret@db.example/forma?sslmode=verify-full")
	t.Setenv("CORS_ORIGINS", "https://shop.example")
	t.Setenv("AUTO_MIGRATE", "false")
	t.Setenv("ENABLE_API_DOCS", "false")
	if _, err := NewConfig(); err != nil {
		t.Fatal(err)
	}
	for _, tt := range []struct{ key, value string }{{"DATABASE_URL", "postgres://runtime:forma_local@db.example/forma?sslmode=verify-full"}, {"DATABASE_URL", "postgres://runtime:secret@db.example/forma?sslmode=disable"}, {"CORS_ORIGINS", "*"}, {"CORS_ORIGINS", "http://shop.example"}, {"AUTO_MIGRATE", "true"}} {
		t.Run(tt.key+tt.value, func(t *testing.T) {
			t.Setenv(tt.key, tt.value)
			if _, err := NewConfig(); err == nil {
				t.Fatal("unsafe production configuration accepted")
			}
		})
	}
}
