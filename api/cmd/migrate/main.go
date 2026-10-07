package main

import (
	"context"
	"fmt"
	"forma/api/migrations"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"os"
	"time"
)

func main() {
	if err := run(); err != nil {
		fmt.Fprintln(os.Stderr, "Migration failed:", err)
		os.Exit(1)
	}
}
func run() error {
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	cfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		return fmt.Errorf("invalid database configuration")
	}
	if os.Getenv("APP_ENV") == "production" && (cfg.ConnConfig.TLSConfig == nil || cfg.ConnConfig.TLSConfig.InsecureSkipVerify) {
		return fmt.Errorf("production migration requires verified TLS")
	}
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return err
	}
	defer pool.Close()
	if err = migrations.Apply(ctx, pool); err != nil {
		return err
	}
	if role := os.Getenv("RUNTIME_DB_ROLE"); role != "" {
		identifier := pgx.Identifier{role}.Sanitize()
		tx, err := pool.Begin(ctx)
		if err != nil {
			return err
		}
		defer tx.Rollback(ctx)
		for _, sql := range []string{`REVOKE CREATE ON SCHEMA public FROM PUBLIC`, `GRANT USAGE ON SCHEMA public TO ` + identifier, `GRANT SELECT,INSERT,UPDATE,DELETE ON ALL TABLES IN SCHEMA public TO ` + identifier, `REVOKE ALL ON schema_migrations FROM ` + identifier, `GRANT USAGE,SELECT ON ALL SEQUENCES IN SCHEMA public TO ` + identifier} {
			if _, err = tx.Exec(ctx, sql); err != nil {
				return err
			}
		}
		if err = tx.Commit(ctx); err != nil {
			return err
		}
	}
	fmt.Println("Migrations applied successfully.")
	return nil
}
