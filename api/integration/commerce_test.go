package integration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	ch "forma/api/internal/catalog/adapters/http"
	cp "forma/api/internal/catalog/adapters/postgres"
	ca "forma/api/internal/catalog/application"
	h "forma/api/internal/commerce/adapters/http"
	p "forma/api/internal/commerce/adapters/postgres"
	"forma/api/internal/commerce/adapters/simulated"
	a "forma/api/internal/commerce/application"
	"forma/api/internal/contact"
	"forma/api/internal/platform"
	"forma/api/migrations"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"io"
	"net/http/httptest"
	"os"
	"sync"
	"sync/atomic"
	"testing"
	"time"
)

func TestCustomerOrdersAndStock(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL")
	}
	ctx := context.Background()
	admin, e := pgx.Connect(ctx, url)
	if e != nil {
		t.Fatal(e)
	}
	defer admin.Close(ctx)
	schema := fmt.Sprintf("forma_commerce_test_%d", time.Now().UnixNano())
	if _, e = admin.Exec(ctx, `CREATE SCHEMA `+schema); e != nil {
		t.Fatal(e)
	}
	defer admin.Exec(ctx, `DROP SCHEMA `+schema+` CASCADE`)
	cfg, e := pgxpool.ParseConfig(url)
	if e != nil {
		t.Fatal(e)
	}
	cfg.ConnConfig.RuntimeParams["search_path"] = schema
	pool, e := pgxpool.NewWithConfig(ctx, cfg)
	if e != nil {
		t.Fatal(e)
	}
	defer pool.Close()
	if e = migrations.Apply(ctx, pool); e != nil {
		t.Fatal(e)
	}
	gateway := &testPaymentGateway{outcome: "declined"}
	service := a.NewService(p.NewFactory(pool), gateway, simulated.NewShipping())
	app := platform.NewServer(platform.Config{AllowedOrigins: "http://127.0.0.1:5173"}, zap.NewNop(), pool, ch.NewHandler(ca.NewService(cp.NewFactory(pool))))
	h.NewHandler(service, pool).Register(app)
	contact.NewHandler(contact.NewRepository(pool)).Register(app)
	req := func(method, path, token string, input any, status int) []byte {
		t.Helper()
		body, _ := json.Marshal(input)
		r := httptest.NewRequest(method, path, bytes.NewReader(body))
		r.Header.Set("Content-Type", "application/json")
		if token != "" {
			r.Header.Set("Authorization", "Bearer "+token)
		}
		res, e := app.Test(r, fiber.TestConfig{Timeout: 10 * time.Second})
		if e != nil {
			t.Fatal(e)
		}
		defer res.Body.Close()
		raw, e := io.ReadAll(res.Body)
		if e != nil {
			t.Fatal(e)
		}
		if res.StatusCode != status {
			t.Fatalf("%s %s: status %d expected %d: %s", method, path, res.StatusCode, status, raw)
		}
		return raw
	}
	decode := func(raw []byte, out any) {
		t.Helper()
		if e := json.Unmarshal(raw, out); e != nil {
			t.Fatal(e)
		}
	}
	req("POST", "/api/v1/contact", "", contact.Message{Name: "Cliente Teste", Email: "teste@example.com", Subject: "Produto", Body: "Gostaria de informações."}, 201)
	var messages int
	if err := pool.QueryRow(ctx, `SELECT count(*) FROM contact_messages`).Scan(&messages); err != nil || messages != 1 {
		t.Fatal("contact persistence", err)
	}
	req("POST", "/api/v1/contact", "", contact.Message{}, 400)
	signup := a.AuthInput{Name: "Teste Cliente", Email: "teste@example.com", Password: "SenhaFicticia123!"}
	var auth, other a.AuthResult
	req("POST", "/api/v1/auth/register", "", signup, 202)
	decode(req("POST", "/api/v1/auth/login", "", a.AuthInput{Email: signup.Email, Password: signup.Password}, 200), &auth)
	if _, e = pool.Exec(ctx, `UPDATE customers SET email_verified_at=now() WHERE id=$1`, auth.Customer.ID); e != nil {
		t.Fatal(e)
	}
	var stored string
	if e = pool.QueryRow(ctx, `SELECT password_hash FROM customers WHERE id=$1`, auth.Customer.ID).Scan(&stored); e != nil || stored == signup.Password {
		t.Fatal("password hash")
	}
	req("POST", "/api/v1/auth/register", "", signup, 202)
	req("POST", "/api/v1/auth/login", "", a.AuthInput{Email: signup.Email, Password: "wrong"}, 401)
	decode(req("POST", "/api/v1/auth/login", "", a.AuthInput{Email: signup.Email, Password: signup.Password}, 200), &auth)
	req("GET", "/api/v1/auth/me", auth.Token, nil, 200)
	req("POST", "/api/v1/auth/register", "", a.AuthInput{Name: "Outro Teste", Email: "outro@example.com", Password: signup.Password}, 202)
	decode(req("POST", "/api/v1/auth/login", "", a.AuthInput{Email: "outro@example.com", Password: signup.Password}, 200), &other)
	if _, e = pool.Exec(ctx, `UPDATE customers SET email_verified_at=now() WHERE id=$1`, other.Customer.ID); e != nil {
		t.Fatal(e)
	}
	in := a.OrderInput{Key: uuid.NewString(), Items: []a.Item{{ID: "sofa-arco", Finish: "Linho natural", Quantity: 2}}, Address: a.Address{CEP: "01001-000", Street: "Rua Teste", Number: "10", City: "São Paulo", State: "SP"}}
	var quote a.ShippingQuote
	decode(req("POST", "/api/v1/shipping/quotes", auth.Token, a.ShippingInput{CEP: in.Address.CEP, Items: in.Items}, 200), &quote)
	in.ShippingQuoteID = quote.ID
	req("POST", "/api/v1/shipping/quotes", "", a.ShippingInput{CEP: in.Address.CEP, Items: in.Items}, 401)
	tampered := in
	tampered.Key = uuid.NewString()
	req("POST", "/api/v1/orders", other.Token, tampered, 409)
	tampered.Address.CEP = "88301401"
	req("POST", "/api/v1/orders", auth.Token, tampered, 409)
	missing := in
	missing.ShippingQuoteID = ""
	req("POST", "/api/v1/orders", auth.Token, missing, 400)
	expired := in
	expired.Key = uuid.NewString()
	var expQuote a.ShippingQuote
	decode(req("POST", "/api/v1/shipping/quotes", auth.Token, a.ShippingInput{CEP: in.Address.CEP, Items: in.Items}, 200), &expQuote)
	expired.ShippingQuoteID = expQuote.ID
	if _, e = pool.Exec(ctx, `UPDATE shipping_quotes SET expires_at=now()-interval '1 second' WHERE id=$1`, expQuote.ID); e != nil {
		t.Fatal(e)
	}
	req("POST", "/api/v1/orders", auth.Token, expired, 409)
	req("POST", "/api/v1/orders", "", in, 401)
	req("POST", "/api/v1/orders", auth.Token, map[string]any{"shippingCents": 1}, 400)
	var order, duplicate a.Order
	decode(req("POST", "/api/v1/orders", auth.Token, in, 201), &order)
	decode(req("POST", "/api/v1/orders", auth.Token, in, 201), &duplicate)
	if order.ID != duplicate.ID || order.TotalCents != 2*429000+quote.AmountCents || order.Status != "awaiting_payment" {
		t.Fatal("idempotency/totals")
	}
	if _, e = pool.Exec(ctx, `UPDATE shipping_quotes SET expires_at=now()-interval '1 second' WHERE id=$1`, quote.ID); e != nil {
		t.Fatal(e)
	}
	decode(req("POST", "/api/v1/orders", auth.Token, in, 201), &duplicate)
	if duplicate.ID != order.ID {
		t.Fatal("expired quote must not invalidate idempotent retry")
	}
	stock := func(product, finish string) int {
		t.Helper()
		var n int
		if e := pool.QueryRow(ctx, `SELECT stock FROM product_variants WHERE product_id=$1 AND finish=$2`, product, finish).Scan(&n); e != nil {
			t.Fatal(e)
		}
		return n
	}
	if stock("sofa-arco", "Linho natural") != 3 {
		t.Fatal("duplicate reservation")
	}
	changed := in
	changed.Items = []a.Item{{ID: "sofa-arco", Finish: "Linho natural", Quantity: 1}}
	req("POST", "/api/v1/orders", auth.Token, changed, 409)
	req("GET", "/api/v1/orders/"+order.ID, other.Token, nil, 404)
	req("POST", "/api/v1/orders/"+order.ID+"/cancel", other.Token, map[string]any{}, 404)
	req("POST", "/api/v1/orders/"+order.ID+"/payment", auth.Token, map[string]string{}, 200)
	for i := 0; i < 2; i++ {
		req("POST", "/api/v1/orders/"+order.ID+"/cancel", auth.Token, map[string]any{}, 200)
	}
	if stock("sofa-arco", "Linho natural") != 5 {
		t.Fatal("cancel must restore exactly once")
	}
	req("POST", "/api/v1/orders/"+order.ID+"/payment", auth.Token, map[string]string{}, 409)
	bad := in
	bad.Key = uuid.NewString()
	bad.Items = []a.Item{{ID: "cadeira-elo", Finish: "Carvalho natural", Quantity: 1}, {ID: "sofa-arco", Finish: "Grafite", Quantity: 1}}
	var badQuote a.ShippingQuote
	decode(req("POST", "/api/v1/shipping/quotes", auth.Token, a.ShippingInput{CEP: bad.Address.CEP, Items: bad.Items}, 200), &badQuote)
	bad.ShippingQuoteID = badQuote.ID
	if _, e = pool.Exec(ctx, `UPDATE product_variants SET stock=0 WHERE product_id='sofa-arco' AND finish='Grafite'`); e != nil {
		t.Fatal(e)
	}
	req("POST", "/api/v1/orders", auth.Token, bad, 409)
	if stock("cadeira-elo", "Carvalho natural") != 5 {
		t.Fatal("rollback leaked stock")
	}
	var singleQuote a.ShippingQuote
	decode(req("POST", "/api/v1/shipping/quotes", auth.Token, a.ShippingInput{CEP: in.Address.CEP, Items: []a.Item{{ID: "sofa-arco", Finish: "Linho natural", Quantity: 1}}}, 200), &singleQuote)
	var wg sync.WaitGroup
	ok := make(chan a.Order, 6)
	fail := make(chan error, 6)
	for i := 0; i < 6; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			input := in
			input.Key = uuid.NewString()
			input.ShippingQuoteID = singleQuote.ID
			input.Items = []a.Item{{ID: "sofa-arco", Finish: "Linho natural", Quantity: 1}}
			o, e := service.CreateOrder(ctx, auth.Customer.ID, input)
			if e != nil {
				fail <- e
			} else {
				ok <- o
			}
		}()
	}
	wg.Wait()
	close(ok)
	close(fail)
	if len(ok) != 5 || len(fail) != 1 || stock("sofa-arco", "Linho natural") != 0 {
		t.Fatalf("oversell: %d successful %d failed", len(ok), len(fail))
	}
	for e := range fail {
		if e != a.StockUnavailable {
			t.Fatal(e)
		}
	}
	pendingOrders := []a.Order{}
	for o := range ok {
		pendingOrders = append(pendingOrders, o)
	}
	paid := pendingOrders[0]
	req("POST", "/api/v1/orders/"+paid.ID+"/payment", auth.Token, map[string]string{"outcome": "approved"}, 400)
	gateway.outcome = "approved"
	before := gateway.calls.Load()
	var paymentGroup sync.WaitGroup
	paymentErrors := make(chan error, 5)
	for i := 0; i < 5; i++ {
		paymentGroup.Add(1)
		go func() {
			defer paymentGroup.Done()
			o, e := service.Pay(ctx, auth.Customer.ID, paid.ID)
			if e != nil {
				paymentErrors <- e
			} else if o.Status != "paid" {
				paymentErrors <- fmt.Errorf("not paid")
			}
		}()
	}
	paymentGroup.Wait()
	close(paymentErrors)
	for err := range paymentErrors {
		t.Fatal(err)
	}
	if gateway.calls.Load() != before+1 {
		t.Fatal("concurrent payment called provider more than once")
	}
	req("POST", "/api/v1/orders/"+paid.ID+"/payment", auth.Token, map[string]string{}, 200)
	req("POST", "/api/v1/orders/"+paid.ID+"/payment", other.Token, map[string]string{}, 404)
	if _, e = pool.Exec(ctx, `UPDATE orders SET reservation_expires_at=now()-interval '1 second' WHERE status='awaiting_payment'`); e != nil {
		t.Fatal(e)
	}
	if _, e = service.Pay(ctx, auth.Customer.ID, pendingOrders[1].ID); e != a.Conflict {
		t.Fatal("expired reservation allowed payment", e)
	}
	for i := 0; i < 2; i++ {
		if err := p.ExpireReservations(ctx, pool); err != nil {
			t.Fatal(err)
		}
	}
	if stock("sofa-arco", "Linho natural") != 4 {
		t.Fatal("reservation expiration must restore stock exactly once")
	}
	identity := p.NewFactory(pool).Identity()
	var attempts sync.WaitGroup
	var admitted atomic.Int64
	for i := 0; i < 24; i++ {
		attempts.Add(1)
		go func() {
			defer attempts.Done()
			allowed, err := identity.AllowAuthentication(ctx, "rate-limit-test")
			if err != nil {
				t.Error(err)
			}
			if allowed {
				admitted.Add(1)
			}
		}()
	}
	attempts.Wait()
	if admitted.Load() != 8 {
		t.Fatal("shared account throttle", admitted.Load())
	}
	req("POST", "/api/v1/orders/"+paid.ID+"/cancel", auth.Token, map[string]any{}, 409)
	req("POST", "/api/v1/orders/"+paid.ID+"/payment", auth.Token, map[string]string{}, 200)
	req("GET", "/api/v1/orders", auth.Token, nil, 200)
	req("GET", "/api/v1/products/sofa-arco/variants", "", nil, 200)
	req("POST", "/api/v1/auth/logout", auth.Token, map[string]any{}, 200)
	req("GET", "/api/v1/auth/me", auth.Token, nil, 401)
	if _, e = pool.Exec(ctx, `UPDATE customer_sessions SET expires_at=now()-interval '1 second' WHERE customer_id=$1`, other.Customer.ID); e != nil {
		t.Fatal(e)
	}
	req("GET", "/api/v1/auth/me", other.Token, nil, 401)
}

type testPaymentGateway struct {
	outcome string
	calls   atomic.Int64
}

func (g *testPaymentGateway) Process(_ context.Context, _ a.Order) (string, error) {
	g.calls.Add(1)
	time.Sleep(20 * time.Millisecond)
	return g.outcome, nil
}
