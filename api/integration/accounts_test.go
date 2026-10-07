package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"forma/api/internal/accounts"
	ch "forma/api/internal/catalog/adapters/http"
	cp "forma/api/internal/catalog/adapters/postgres"
	ca "forma/api/internal/catalog/application"
	h "forma/api/internal/commerce/adapters/http"
	p "forma/api/internal/commerce/adapters/postgres"
	"forma/api/internal/commerce/adapters/simulated"
	a "forma/api/internal/commerce/application"
	"forma/api/internal/platform"
	"forma/api/internal/traffic"
	"forma/api/migrations"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"io"
	"net/http/httptest"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestAccountLifecycle(t *testing.T) {
	database := os.Getenv("TEST_DATABASE_URL")
	if database == "" {
		t.Skip("set TEST_DATABASE_URL")
	}
	ctx := context.Background()
	admin, err := pgx.Connect(ctx, database)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(ctx)
	schema := fmt.Sprintf("forma_accounts_test_%d", time.Now().UnixNano())
	if _, err = admin.Exec(ctx, `CREATE SCHEMA `+schema); err != nil {
		t.Fatal(err)
	}
	defer admin.Exec(ctx, `DROP SCHEMA `+schema+` CASCADE`)
	cfg, err := pgxpool.ParseConfig(database)
	if err != nil {
		t.Fatal(err)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		t.Fatal(err)
	}
	defer pool.Close()
	if err = migrations.Apply(ctx, pool); err != nil {
		t.Fatal(err)
	}
	mailCfg := accounts.MailConfig{Origin: "http://127.0.0.1:5173", Key: bytes.Repeat([]byte{7}, 32)}
	lifecycle := accounts.NewService(pool, mailCfg)
	service := a.NewService(p.NewFactory(pool), &testPaymentGateway{outcome: "approved"}, simulated.NewShipping())
	service.SetRegistrationPreparer(lifecycle.Prepare)
	app := platform.NewServer(platform.Config{AllowedOrigins: "http://127.0.0.1:5173"}, zap.NewNop(), pool, ch.NewHandler(ca.NewService(cp.NewFactory(pool))))
	h.NewHandler(service, pool).Register(app)
	accounts.NewHandler(lifecycle).Register(app)
	request := func(method, path, token string, body any, status int) []byte {
		t.Helper()
		raw, _ := json.Marshal(body)
		r := httptest.NewRequest(method, path, bytes.NewReader(raw))
		r.Header.Set("Content-Type", "application/json")
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		response, err := app.Test(r, fiber.TestConfig{Timeout: 10 * time.Second})
		if err != nil {
			t.Fatal(err)
		}
		defer response.Body.Close()
		result, _ := io.ReadAll(response.Body)
		if response.StatusCode != status {
			t.Fatalf("%s: expected %d got %d %s", path, status, response.StatusCode, result)
		}
		return result
	}
	input := a.AuthInput{Name: "Cliente Segurança", Email: "security@example.com", Password: "PasswordForTests123!"}
	first := request("POST", "/api/v1/auth/register", "", input, 202)
	second := request("POST", "/api/v1/auth/register", "", input, 202)
	if !bytes.Equal(first, second) {
		t.Fatal("registration enumerates accounts")
	}
	challenge := func(subject string) string {
		t.Helper()
		var encrypted string
		if err := pool.QueryRow(ctx, `SELECT body FROM email_outbox WHERE subject=$1 ORDER BY created_at DESC LIMIT 1`, subject).Scan(&encrypted); err != nil {
			t.Fatal(err)
		}
		body, err := accounts.DecodeBody(mailCfg.Key, encrypted)
		if err != nil {
			t.Fatal(err)
		}
		i := strings.Index(body, "#token=")
		if i < 0 {
			t.Fatal("missing token link")
		}
		token := strings.Split(body[i+7:], "\n")[0]
		if len(token) != 43 || strings.Contains(encrypted, token) {
			t.Fatal("token must be encrypted in outbox")
		}
		return token
	}
	login := func(password string) a.AuthResult {
		t.Helper()
		var result a.AuthResult
		raw := request("POST", "/api/v1/auth/login", "", a.AuthInput{Email: input.Email, Password: password}, 200)
		if err := json.Unmarshal(raw, &result); err != nil {
			t.Fatal(err)
		}
		return result
	}
	auth := login(input.Password)
	if auth.Customer.EmailVerified {
		t.Fatal("unverified registration")
	}
	request("POST", "/api/v1/auth/profile", "", map[string]string{"name": "Outro nome"}, 401)
	request("POST", "/api/v1/auth/profile", auth.Token, map[string]string{"name": " "}, 400)
	request("POST", "/api/v1/auth/profile", auth.Token, map[string]string{"name": strings.Repeat("x", 101)}, 400)
	request("POST", "/api/v1/auth/profile", auth.Token, map[string]string{"name": "Nome\nInválido"}, 400)
	request("POST", "/api/v1/auth/profile", auth.Token, map[string]string{"name": "Nome válido", "email": "other@example.com"}, 400)
	request("POST", "/api/v1/auth/profile", auth.Token, map[string]string{"name": "  João Atualizado  "}, 200)
	var updated a.Customer
	if err := json.Unmarshal(request("GET", "/api/v1/auth/me", auth.Token, nil, 200), &updated); err != nil || updated.Name != "João Atualizado" || updated.Email != input.Email || updated.EmailVerified || updated.ID != auth.Customer.ID {
		t.Fatal("profile update did not persist safely", err)
	}
	request("POST", "/api/v1/orders", auth.Token, map[string]any{}, 403)
	token := challenge("Confirme seu e-mail na Forma")
	request("POST", "/api/v1/auth/verify/confirm", "", map[string]string{"token": token}, 200)
	request("POST", "/api/v1/auth/verify/confirm", "", map[string]string{"token": token}, 400)
	var customer a.Customer
	if err := json.Unmarshal(request("GET", "/api/v1/auth/me", auth.Token, nil, 200), &customer); err != nil || !customer.EmailVerified {
		t.Fatal("verification failed", err)
	}
	otherSession := login(input.Password)
	request("POST", "/api/v1/auth/sessions/revoke", auth.Token, map[string]any{}, 200)
	request("GET", "/api/v1/auth/me", otherSession.Token, nil, 401)
	request("POST", "/api/v1/auth/profile", otherSession.Token, map[string]string{"name": "Revoked update"}, 401)
	request("GET", "/api/v1/auth/me", auth.Token, nil, 200)
	known := request("POST", "/api/v1/auth/reset/request", "", map[string]string{"email": input.Email}, 200)
	unknown := request("POST", "/api/v1/auth/reset/request", "", map[string]string{"email": "missing@example.com"}, 200)
	if !bytes.Equal(known, unknown) {
		t.Fatal("reset request enumerates accounts")
	}
	reset := challenge("Redefina sua senha na Forma")
	request("POST", "/api/v1/auth/verify/confirm", "", map[string]string{"token": reset}, 400)
	// Race consumption at the application boundary: only one reset may succeed.
	var success atomic.Int64
	var wg sync.WaitGroup
	for i := 0; i < 2; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			if err := lifecycle.Consume(ctx, reset, "reset", "ReplacementPassword123!"); err == nil {
				success.Add(1)
			}
		}()
	}
	wg.Wait()
	if success.Load() != 1 {
		t.Fatalf("single-use reset: %d successes", success.Load())
	}
	request("GET", "/api/v1/auth/me", auth.Token, nil, 401)
	request("POST", "/api/v1/auth/login", "", a.AuthInput{Email: input.Email, Password: input.Password}, 401)
	newAuth := login("ReplacementPassword123!")
	request("GET", "/api/v1/auth/me", newAuth.Token, nil, 200)
	request("POST", "/api/v1/auth/reset/confirm", "", map[string]string{"token": reset, "password": "ReplacementPassword123!"}, 400)
	var events int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM security_events`).Scan(&events); err != nil || events != 3 {
		t.Fatal("security events", events, err)
	}
	// Shared limit: independent concurrent calls cannot bypass the bucket.
	var allowed atomic.Int64
	for i := 0; i < 32; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ok, err := traffic.Allow(ctx, pool, "concurrent-scope", 8, time.Minute)
			if err != nil {
				t.Error(err)
			}
			if ok {
				allowed.Add(1)
			}
		}()
	}
	wg.Wait()
	if allowed.Load() != 8 {
		t.Fatalf("shared limiter allowed %d", allowed.Load())
	}
	// An expired challenge cannot change the password.
	if _, err = pool.Exec(ctx, `UPDATE account_challenges SET created_at=now()-interval '2 minutes'`); err != nil {
		t.Fatal(err)
	}
	request("POST", "/api/v1/auth/reset/request", "", map[string]string{"email": input.Email}, 200)
	expired := challenge("Redefina sua senha na Forma")
	if _, err = pool.Exec(ctx, `UPDATE account_challenges SET expires_at=now()-interval '1 second'`); err != nil {
		t.Fatal(err)
	}
	request("POST", "/api/v1/auth/reset/confirm", "", map[string]string{"token": expired, "password": "AnotherNewPassword123!"}, 400)
}
