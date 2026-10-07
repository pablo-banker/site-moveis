package traffic

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	a "forma/api/internal/commerce/application"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"net"
	"strconv"
	"strings"
	"time"
)

func Allow(ctx context.Context, pool *pgxpool.Pool, key string, max int, window time.Duration) (bool, error) {
	hash := sha256.Sum256([]byte(key))
	var count int
	err := pool.QueryRow(ctx, `INSERT INTO rate_buckets(key_hash,attempts,resets_at)VALUES($1,1,clock_timestamp()+$2::interval)
 ON CONFLICT(key_hash)DO UPDATE SET attempts=CASE WHEN rate_buckets.resets_at<=clock_timestamp() THEN 1 ELSE rate_buckets.attempts+1 END,
 resets_at=CASE WHEN rate_buckets.resets_at<=clock_timestamp() THEN clock_timestamp()+$2::interval ELSE rate_buckets.resets_at END
 WHERE rate_buckets.resets_at<=clock_timestamp() OR rate_buckets.attempts<$3 RETURNING attempts`, fmt.Sprintf("%x", hash), fmt.Sprintf("%f seconds", window.Seconds()), max).Scan(&count)
	if err == pgx.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}
func Client(c fiber.Ctx, secret string, required bool) (string, error) {
	if secret == "" {
		if required {
			return "", fmt.Errorf("BFF secret missing")
		}
		return c.IP(), nil
	}
	stamp := c.Get("X-Forma-Timestamp")
	at, err := strconv.ParseInt(stamp, 10, 64)
	if err != nil || time.Now().Unix()-at > 30 || at-time.Now().Unix() > 30 {
		return "", fiber.ErrForbidden
	}
	client := c.Get("X-Forma-Client")
	if net.ParseIP(client) == nil {
		return "", fiber.ErrForbidden
	}
	mac := hmac.New(sha256.New, []byte(secret))
	_, _ = mac.Write([]byte(stamp + "\n" + c.Method() + "\n" + c.Path() + "\n" + client))
	provided, err := hex.DecodeString(c.Get("X-Forma-Signature"))
	if err != nil || !hmac.Equal(provided, mac.Sum(nil)) {
		return "", fiber.ErrForbidden
	}
	return client, nil
}
func Middleware(pool *pgxpool.Pool, secret string, production bool) fiber.Handler {
	return func(c fiber.Ctx) error {
		client, err := Client(c, secret, production)
		if err != nil {
			return err
		}
		scope, max := "api", 120
		path := c.Path()
		if strings.HasPrefix(path, "/api/v1/auth/") {
			scope, max = "auth", 30
		}
		if strings.HasPrefix(path, "/api/v1/addresses/") {
			scope, max = "address", 30
		}
		if path == "/api/v1/contact" {
			scope, max = "contact", 5
		}
		if c.Method() == "POST" && strings.HasPrefix(path, "/api/v1/orders") {
			scope, max = "order", 20
			if strings.HasSuffix(path, "/payment") {
				scope = "payment"
			}
			if strings.HasSuffix(path, "/cancel") {
				scope = "cancel"
			}
		}
		if path == "/api/v1/shipping/quotes" {
			scope, max = "shipping", 30
		}
		ctx, cancel := context.WithTimeout(c.Context(), 2*time.Second)
		defer cancel()
		allowed, err := Allow(ctx, pool, scope+":"+client, max, time.Minute)
		if err != nil {
			return err
		}
		if !allowed {
			c.Set("Retry-After", "60")
			return &a.Failure{Status: 429, Code: "rate_limited", Message: "Muitas solicitações. Aguarde um minuto."}
		}
		return c.Next()
	}
}
