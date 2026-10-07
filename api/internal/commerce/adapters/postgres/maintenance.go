package postgres

import (
	"context"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"time"
)

// Row locks and SKIP LOCKED allow multiple workers without double restoration.
func ExpireReservations(ctx context.Context, pool *pgxpool.Pool) error {
	tx, err := pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	rows, err := tx.Query(ctx, `SELECT id::text,customer_id::text FROM orders WHERE status='awaiting_payment' AND reservation_expires_at<=now() ORDER BY reservation_expires_at LIMIT 20 FOR UPDATE SKIP LOCKED`)
	if err != nil {
		return err
	}
	type target struct{ id, user string }
	targets := []target{}
	for rows.Next() {
		var t target
		if err := rows.Scan(&t.id, &t.user); err != nil {
			rows.Close()
			return err
		}
		targets = append(targets, t)
	}
	rows.Close()
	if err := rows.Err(); err != nil {
		return err
	}
	for _, t := range targets {
		order, err := read(ctx, tx, t.user, t.id)
		if err != nil {
			return err
		}
		for _, item := range order.Items {
			if _, err := tx.Exec(ctx, `UPDATE product_variants SET stock=stock+$3 WHERE product_id=$1 AND finish=$2`, item.ID, item.Finish, item.Quantity); err != nil {
				return err
			}
		}
		if _, err := tx.Exec(ctx, `UPDATE orders SET status='cancelled' WHERE id=$1`, t.id); err != nil {
			return err
		}
	}
	return tx.Commit(ctx)
}

func RunMaintenance(lc fx.Lifecycle, pool *pgxpool.Pool, log *zap.Logger) {
	var cancel context.CancelFunc
	done := make(chan struct{})
	lc.Append(fx.Hook{OnStart: func(context.Context) error {
		ctx, stop := context.WithCancel(context.Background())
		cancel = stop
		go func() {
			defer close(done)
			ticker := time.NewTicker(time.Minute)
			defer ticker.Stop()
			for {
				work, stop := context.WithTimeout(ctx, 10*time.Second)
				if err := ExpireReservations(work, pool); err != nil && ctx.Err() == nil {
					log.Error("reservation maintenance failed", zap.Error(err))
				}
				// Bounded garbage collection; order snapshots do not depend on expired quotes.
				for _, query := range []string{
					`DELETE FROM customer_sessions WHERE token_hash IN (SELECT token_hash FROM customer_sessions WHERE expires_at<now()-interval '1 day' LIMIT 500)`,
					`DELETE FROM shipping_quotes WHERE id IN (SELECT id FROM shipping_quotes WHERE expires_at<now()-interval '1 day' LIMIT 500)`,
					`DELETE FROM account_challenges WHERE token_hash IN (SELECT token_hash FROM account_challenges WHERE expires_at<now() LIMIT 500)`,
					`DELETE FROM email_outbox WHERE id IN (SELECT id FROM email_outbox WHERE created_at<now()-interval '7 days' LIMIT 500)`,
					`DELETE FROM security_events WHERE id IN (SELECT id FROM security_events WHERE created_at<now()-interval '90 days' LIMIT 500)`,
					`DELETE FROM rate_buckets WHERE key_hash IN (SELECT key_hash FROM rate_buckets WHERE resets_at<now()-interval '1 day' LIMIT 500)`,
					`DELETE FROM address_cache WHERE cep IN (SELECT cep FROM address_cache WHERE expires_at<now()-interval '1 day' LIMIT 500)`,
					`DELETE FROM auth_attempts WHERE key_hash IN (SELECT key_hash FROM auth_attempts WHERE resets_at<now()-interval '1 day' LIMIT 500)`,
				} {
					if _, err := pool.Exec(work, query); err != nil && ctx.Err() == nil {
						log.Error("maintenance cleanup failed", zap.Error(err))
					}
				}
				stop()
				select {
				case <-ctx.Done():
					return
				case <-ticker.C:
				}
			}
		}()
		return nil
	}, OnStop: func(ctx context.Context) error {
		if cancel != nil {
			cancel()
		}
		select {
		case <-done:
			return nil
		case <-ctx.Done():
			return ctx.Err()
		}
	}})
}
