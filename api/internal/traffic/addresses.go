package traffic

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"net/http"
	"regexp"
	"time"
)

var validCEP = regexp.MustCompile(`^\d{8}$`)

func AddressHandler(pool *pgxpool.Pool) fiber.Handler {
	slots := make(chan struct{}, 8)
	client := &http.Client{Timeout: 5 * time.Second, CheckRedirect: func(*http.Request, []*http.Request) error { return fmt.Errorf("redirect not permitted") }}
	return func(c fiber.Ctx) error {
		code := c.Params("cep")
		if !validCEP.MatchString(code) {
			return fiber.NewError(400, "Informe um CEP com 8 dígitos.")
		}
		ctx, cancel := context.WithTimeout(c.Context(), 6*time.Second)
		defer cancel()
		var cached json.RawMessage
		if err := pool.QueryRow(ctx, `SELECT address FROM address_cache WHERE cep=$1 AND expires_at>now()`, code).Scan(&cached); err == nil {
			c.Set("Cache-Control", "public, max-age=3600")
			return c.JSON(cached)
		}
		select {
		case slots <- struct{}{}:
			defer func() { <-slots }()
		default:
			return fiber.NewError(503, "Consulta de CEP indisponível. Preencha o endereço manualmente.")
		}
		req, err := http.NewRequestWithContext(ctx, "GET", "https://viacep.com.br/ws/"+code+"/json/", nil)
		if err != nil {
			return err
		}
		response, err := client.Do(req)
		if err != nil {
			return fiber.NewError(503, "Não foi possível consultar o CEP.")
		}
		defer response.Body.Close()
		if response.StatusCode != 200 {
			return fiber.NewError(503, "Não foi possível consultar o CEP.")
		}
		raw, err := io.ReadAll(io.LimitReader(response.Body, 32769))
		if err != nil || len(raw) > 32768 {
			return fiber.NewError(503, "Não foi possível consultar o CEP.")
		}
		var v struct {
			Street string `json:"logradouro"`
			City   string `json:"localidade"`
			State  string `json:"uf"`
			Error  bool   `json:"erro"`
		}
		if json.Unmarshal(raw, &v) != nil {
			return fiber.NewError(503, "Não foi possível consultar o CEP.")
		}
		if v.Error {
			return fiber.NewError(404, "CEP não encontrado.")
		}
		if v.City == "" || len(v.City) > 100 || len(v.Street) > 200 || len(v.State) != 2 {
			return fiber.NewError(503, "Não foi possível consultar o CEP.")
		}
		address := fiber.Map{"street": v.Street, "city": v.City, "state": v.State}
		encoded, _ := json.Marshal(address)
		if _, err = pool.Exec(ctx, `INSERT INTO address_cache(cep,address,expires_at)VALUES($1,$2,now()+interval '1 hour')ON CONFLICT(cep)DO UPDATE SET address=excluded.address,expires_at=excluded.expires_at`, code, encoded); err != nil {
			return err
		}
		c.Set("Cache-Control", "public, max-age=3600")
		return c.JSON(address)
	}
}
