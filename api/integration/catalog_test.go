package integration_test

import (
	"context"
	"encoding/json"
	"fmt"
	"forma/api/docs"
	cataloghttp "forma/api/internal/catalog/adapters/http"
	"forma/api/internal/catalog/adapters/postgres"
	"forma/api/internal/catalog/application"
	"forma/api/internal/platform"
	"forma/api/migrations"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
	"io"
	"net/http/httptest"
	"os"
	"testing"
	"time"
)

func TestCatalogPostgresHTTP(t *testing.T) {
	url := os.Getenv("TEST_DATABASE_URL")
	if url == "" {
		t.Skip("set TEST_DATABASE_URL to run isolated Postgres integration tests")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	admin, err := pgx.Connect(ctx, url)
	if err != nil {
		t.Fatal(err)
	}
	defer admin.Close(context.Background())
	schema := fmt.Sprintf("forma_test_%d", time.Now().UnixNano())
	if _, err = admin.Exec(ctx, `CREATE SCHEMA `+schema); err != nil {
		t.Fatal(err)
	}
	defer func() {
		if _, err := admin.Exec(context.Background(), `DROP SCHEMA `+schema+` CASCADE`); err != nil {
			t.Error(err)
		}
	}()
	cfg, err := pgxpool.ParseConfig(url)
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
	if err = migrations.Apply(ctx, pool); err != nil {
		t.Fatalf("idempotent migration: %v", err)
	}
	factory := postgres.NewFactory(pool)
	service := application.NewService(factory)
	app := platform.NewServer(platform.Config{EnableDocs: true, AllowedOrigins: "http://127.0.0.1:5173"}, zap.NewNop(), pool, cataloghttp.NewHandler(service))
	request := func(path string, status int) []byte {
		t.Helper()
		res, err := app.Test(httptest.NewRequest("GET", path, nil), fiber.TestConfig{Timeout: 5 * time.Second})
		if err != nil {
			t.Fatal(err)
		}
		defer res.Body.Close()
		body, err := io.ReadAll(res.Body)
		if err != nil {
			t.Fatal(err)
		}
		if res.StatusCode != status {
			t.Fatalf("%s: status=%d expected=%d body=%s", path, res.StatusCode, status, body)
		}
		if res.Header.Get("X-Request-ID") == "" {
			t.Fatal("missing request ID")
		}
		if !json.Valid(body) {
			t.Fatalf("invalid JSON: %s", body)
		}
		return body
	}
	type page struct {
		Data  []cataloghttp.Product `json:"data"`
		Total int                   `json:"total"`
	}
	decode := func(path string) page {
		t.Helper()
		var p page
		if err := json.Unmarshal(request(path, 200), &p); err != nil {
			t.Fatal(err)
		}
		return p
	}
	all := decode("/api/v1/products")
	if all.Total != 8 || len(all.Data) != 8 {
		t.Fatalf("seed catalog: %+v", all)
	}
	byRoom := decode("/api/v1/products?room=sala-de-estar&sort=price_asc&limit=2&offset=1")
	if byRoom.Total != 4 || len(byRoom.Data) != 2 || byRoom.Data[0].PriceCents > byRoom.Data[1].PriceCents {
		t.Fatalf("filter/sort/page: %+v", byRoom)
	}
	search := decode("/api/v1/products?q=Arco")
	if search.Total != 1 || search.Data[0].ID != "sofa-arco" {
		t.Fatalf("search: %+v", search)
	}
	categories := decode("/api/v1/products?category=cadeiras")
	if categories.Total != 1 {
		t.Fatal("category filter")
	}
	empty := decode("/api/v1/products?offset=100")
	if len(empty.Data) != 0 || empty.Total != 8 {
		t.Fatal("out-of-range pagination")
	}
	injection := decode("/api/v1/products?q=%27%20OR%201%3D1--")
	if injection.Total != 0 {
		t.Fatal("query must be a literal parameter")
	}
	var product cataloghttp.Product
	if err := json.Unmarshal(request("/api/v1/products/sofa-arco", 200), &product); err != nil {
		t.Fatal(err)
	}
	if product.PriceCents != 429000 || product.Price != 4290 || len(product.Finishes) != 3 {
		t.Fatalf("product DTO: %+v", product)
	}
	for _, path := range []string{"/api/v1/products?limit=0", "/api/v1/products?limit=101", "/api/v1/products?offset=-1", "/api/v1/products?limit=abc", "/api/v1/products?sort=invalid"} {
		request(path, 400)
	}
	request("/api/v1/products/missing", 404)
	request("/missing", 404)
	for _, path := range []string{"/api/v1/categories", "/api/v1/rooms", "/health/live", "/health/ready", "/openapi.json"} {
		request(path, 200)
	}
	// Updating the database must change API output, not depend on the demo source.
	if _, err := pool.Exec(ctx, `UPDATE products SET price_cents=430099 WHERE id='sofa-arco'`); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(request("/api/v1/products/sofa-arco", 200), &product); err != nil {
		t.Fatal(err)
	}
	if product.Price != 4300.99 {
		t.Fatal("API did not use persisted price")
	}
	// Applied migrations never overwrite existing data.
	if err = migrations.Apply(ctx, pool); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(request("/api/v1/products/sofa-arco", 200), &product); err != nil {
		t.Fatal(err)
	}
	if product.PriceCents != 430099 {
		t.Fatal("migration overwrote persisted changes")
	}
}
func TestOpenAPIContract(t *testing.T) {
	var spec struct {
		Paths map[string]json.RawMessage `json:"paths"`
	}
	if err := json.Unmarshal(docs.Spec, &spec); err != nil {
		t.Fatal(err)
	}
	for _, path := range []string{"/api/v1/products", "/api/v1/products/{id}", "/api/v1/categories", "/api/v1/rooms", "/health/live", "/health/ready"} {
		if _, ok := spec.Paths[path]; !ok {
			t.Fatalf("undocumented route %s", path)
		}
	}
}
