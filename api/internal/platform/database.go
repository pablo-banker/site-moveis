package platform

import (
	"context"
	"fmt"
	"forma/api/migrations"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"time"
)

func NewDatabase(lc fx.Lifecycle, c Config) (*pgxpool.Pool, error) {
	cfg, err := pgxpool.ParseConfig(c.DatabaseURL)
	if err != nil {
		return nil, err
	}
	cfg.MaxConns = 10
	cfg.ConnConfig.ConnectTimeout = 5 * time.Second
	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{OnStart: func(ctx context.Context) error {
		if err := pool.Ping(ctx); err != nil {
			pool.Close()
			return err
		}
		if c.Environment == "production" {
			var privileged bool
			if err := pool.QueryRow(ctx, `SELECT rolsuper OR rolcreatedb OR rolcreaterole OR has_schema_privilege(current_user,'public','CREATE') OR current_user=(SELECT pg_get_userbyid(datdba) FROM pg_database WHERE datname=current_database()) FROM pg_roles WHERE rolname=current_user`).Scan(&privileged); err != nil {
				pool.Close()
				return err
			}
			if privileged {
				pool.Close()
				return fmt.Errorf("production runtime database role must not have administrative, ownership or schema CREATE privileges")
			}
		}
		if c.AutoMigrate {
			if err := migrations.Apply(ctx, pool); err != nil {
				pool.Close()
				return err
			}
		}
		return nil
	}, OnStop: func(context.Context) error { pool.Close(); return nil }})
	return pool, nil
}
