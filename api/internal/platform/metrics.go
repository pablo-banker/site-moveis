package platform

import (
	"context"
	"crypto/subtle"
	"errors"
	"fmt"
	a "forma/api/internal/commerce/application"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"strings"
	"sync/atomic"
	"time"
)

type metrics struct{ requests, failures, nanos atomic.Uint64 }

func (m *metrics) observe(c fiber.Ctx) error {
	start := time.Now()
	err := c.Next()
	m.requests.Add(1)
	m.nanos.Add(uint64(time.Since(start)))
	status := c.Response().StatusCode()
	if err != nil {
		status = 500
		var fe *fiber.Error
		var ce *a.Failure
		if errors.As(err, &fe) {
			status = fe.Code
		}
		if errors.As(err, &ce) {
			status = ce.Status
		}
	}
	if status >= 500 {
		m.failures.Add(1)
	}
	return err
}
func (m *metrics) handler(pool *pgxpool.Pool, token string) fiber.Handler {
	return func(c fiber.Ctx) error {
		if !strings.HasPrefix(c.Get("Authorization"), "Bearer ") {
			return fiber.ErrNotFound
		}
		provided := strings.TrimPrefix(c.Get("Authorization"), "Bearer ")
		if token == "" || subtle.ConstantTimeCompare([]byte(provided), []byte(token)) != 1 {
			return fiber.ErrNotFound
		}
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()
		var pending, dead, expired int64
		var oldest float64
		err := pool.QueryRow(ctx, `SELECT count(*) FILTER(WHERE sent_at IS NULL AND attempts<10),count(*) FILTER(WHERE sent_at IS NULL AND attempts>=10),COALESCE(EXTRACT(EPOCH FROM now()-min(created_at) FILTER(WHERE sent_at IS NULL AND attempts<10)),0) FROM email_outbox`).Scan(&pending, &dead, &oldest)
		if err != nil {
			return err
		}
		if err = pool.QueryRow(ctx, `SELECT count(*) FROM orders WHERE status='awaiting_payment' AND reservation_expires_at<now()-interval '5 minutes'`).Scan(&expired); err != nil {
			return err
		}
		stats := pool.Stat()
		c.Set("Content-Type", "text/plain; version=0.0.4")
		c.Set("Cache-Control", "no-store")
		return c.SendString(fmt.Sprintf("forma_http_requests_total %d\nforma_http_errors_total %d\nforma_http_duration_seconds_total %f\nforma_email_pending %d\nforma_email_failed %d\nforma_email_oldest_seconds %f\nforma_expired_reservations %d\nforma_database_connections %d\nforma_database_max_connections %d\n", m.requests.Load(), m.failures.Load(), float64(m.nanos.Load())/float64(time.Second), pending, dead, oldest, expired, stats.AcquiredConns(), stats.MaxConns()))
	}
}
