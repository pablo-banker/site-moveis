package accounts

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	a "forma/api/internal/commerce/application"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"net/mail"
	"strings"
	"time"
	"unicode/utf8"
)

type Service struct {
	pool   *pgxpool.Pool
	origin string
	slots  chan struct{}
	key    []byte
}

func NewService(pool *pgxpool.Pool, cfg MailConfig) *Service {
	return &Service{pool: pool, origin: cfg.Origin, slots: a.PasswordHashSlots(), key: cfg.Key}
}
func Hash(token string) string { return fmt.Sprintf("%x", sha256.Sum256([]byte(token))) }
func (s *Service) Prepare(c a.Customer) (*a.AccountDelivery, error) {
	return s.delivery(c.Email, "verify")
}

// Same response for absent/verified accounts; tokens are never returned through HTTP.
func (s *Service) issue(ctx context.Context, email, purpose string) error {
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var id string
	var verified bool
	err = tx.QueryRow(ctx, `SELECT id::text,email_verified_at IS NOT NULL FROM customers WHERE email=$1 FOR UPDATE`, email).Scan(&id, &verified)
	if err == pgx.ErrNoRows {
		return nil
	}
	if err != nil {
		return err
	}
	if purpose == "verify" && verified {
		return nil
	}
	// Only one delivery per account/purpose/minute, also across API replicas.
	var recent bool
	if err = tx.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM account_challenges WHERE customer_id=$1 AND purpose=$2 AND created_at>now()-interval '1 minute')`, id, purpose).Scan(&recent); err != nil {
		return err
	}
	if recent {
		return nil
	}
	delivery, err := s.delivery(email, purpose)
	if err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `DELETE FROM account_challenges WHERE customer_id=$1 AND purpose=$2`, id, purpose); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO account_challenges(token_hash,customer_id,purpose,expires_at)VALUES($1,$2,$3,$4)`, delivery.TokenHash, id, purpose, delivery.ExpiresAt); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO email_outbox(id,recipient,subject,body)VALUES($1,$2,$3,$4)`, uuid.NewString(), delivery.Email, delivery.Subject, delivery.Body); err != nil {
		return err
	}
	return tx.Commit(ctx)
}
func (s *Service) Consume(ctx context.Context, token, purpose, password string) error {
	if len(token) != 43 {
		return a.Invalid("Link inválido ou vencido. Solicite um novo link.")
	}
	var hash string
	var err error
	if purpose == "reset" {
		if utf8.RuneCountInString(password) < 15 || len(password) > 128 {
			return a.Invalid("Use uma senha de pelo menos 15 caracteres (máximo 128 bytes).")
		}
		select {
		case s.slots <- struct{}{}:
			defer func() { <-s.slots }()
		default:
			return &a.Failure{Status: 503, Code: "busy", Message: "Tente novamente em instantes."}
		}
		// Check before expensive hashing; the transaction below still guarantees single use.
		var valid bool
		if err = s.pool.QueryRow(ctx, `SELECT EXISTS(SELECT 1 FROM account_challenges WHERE token_hash=$1 AND purpose=$2 AND expires_at>clock_timestamp())`, Hash(token), purpose).Scan(&valid); err != nil {
			return err
		}
		if !valid {
			return a.Invalid("Link inválido ou vencido. Solicite um novo link.")
		}
		hash, err = a.HashPassword(password)
		if err != nil {
			return err
		}
	}
	tx, err := s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	var id string
	// Customer lock serializes issuing/consuming tokens and prevents cross-purpose deadlocks.
	err = tx.QueryRow(ctx, `SELECT c.id::text FROM customers c JOIN account_challenges t ON t.customer_id=c.id WHERE t.token_hash=$1 AND t.purpose=$2 FOR UPDATE OF c`, Hash(token), purpose).Scan(&id)
	if err == pgx.ErrNoRows {
		return a.Invalid("Link inválido ou vencido. Solicite um novo link.")
	}
	if err != nil {
		return err
	}
	var used string
	err = tx.QueryRow(ctx, `DELETE FROM account_challenges WHERE token_hash=$1 AND purpose=$2 AND expires_at>clock_timestamp() RETURNING customer_id::text`, Hash(token), purpose).Scan(&used)
	if err == pgx.ErrNoRows {
		return a.Invalid("Link inválido ou vencido. Solicite um novo link.")
	}
	if err != nil {
		return err
	}
	event := "email_verified"
	if purpose == "reset" {
		event = "password_reset"
		if _, err = tx.Exec(ctx, `UPDATE customers SET password_hash=$2,email_verified_at=COALESCE(email_verified_at,now()) WHERE id=$1`, id, hash); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `DELETE FROM customer_sessions WHERE customer_id=$1`, id); err != nil {
			return err
		}
		if _, err = tx.Exec(ctx, `DELETE FROM account_challenges WHERE customer_id=$1`, id); err != nil {
			return err
		}
	} else {
		if _, err = tx.Exec(ctx, `UPDATE customers SET email_verified_at=COALESCE(email_verified_at,now()) WHERE id=$1`, id); err != nil {
			return err
		}
	}
	if _, err = tx.Exec(ctx, `INSERT INTO security_events(customer_id,event)VALUES($1,$2)`, id, event); err != nil {
		return err
	}
	return tx.Commit(ctx)
}

type Handler struct{ s *Service }

func NewHandler(s *Service) *Handler { return &Handler{s} }
func (h *Handler) Register(app *fiber.App) {
	for _, purpose := range []string{"verify", "reset"} {
		app.Post("/api/v1/auth/"+purpose+"/request", h.request(purpose))
		app.Post("/api/v1/auth/"+purpose+"/confirm", h.confirm(purpose))
	}
	app.Get("/api/v1/auth/sessions", h.sessions)
	app.Post("/api/v1/auth/sessions/revoke", h.revoke)
}
func decode(c fiber.Ctx, v any) error {
	if strings.Split(c.Get("Content-Type"), ";")[0] != "application/json" {
		return &a.Failure{Status: 415, Code: "json_required", Message: "Envie os dados como JSON."}
	}
	d := json.NewDecoder(strings.NewReader(string(c.Body())))
	d.DisallowUnknownFields()
	if d.Decode(v) != nil || d.Decode(new(any)) != io.EOF {
		return a.Invalid("Confira os dados enviados.")
	}
	return nil
}
func (h *Handler) request(purpose string) fiber.Handler {
	return func(c fiber.Ctx) error {
		var in struct {
			Email string `json:"email"`
		}
		if err := decode(c, &in); err != nil {
			return err
		}
		in.Email = strings.ToLower(strings.TrimSpace(in.Email))
		parsed, err := mail.ParseAddress(in.Email)
		if err != nil || parsed.Address != in.Email || len(in.Email) > 254 {
			return a.Invalid("Informe um e-mail válido.")
		}
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()
		if err = h.s.issue(ctx, in.Email, purpose); err != nil {
			return err
		}
		return c.JSON(fiber.Map{"message": "Se o endereço estiver cadastrado, você receberá um e-mail com as instruções."})
	}
}
func (h *Handler) confirm(purpose string) fiber.Handler {
	return func(c fiber.Ctx) error {
		var in struct {
			Token    string `json:"token"`
			Password string `json:"password"`
		}
		if err := decode(c, &in); err != nil {
			return err
		}
		ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
		defer cancel()
		if err := h.s.Consume(ctx, in.Token, purpose, in.Password); err != nil {
			return err
		}
		return c.JSON(fiber.Map{"ok": true})
	}
}
func currentToken(c fiber.Ctx) string {
	v := c.Get("Authorization")
	if strings.HasPrefix(v, "Bearer ") {
		return strings.TrimPrefix(v, "Bearer ")
	}
	return ""
}
func (h *Handler) user(ctx context.Context, c fiber.Ctx) (string, error) {
	token := currentToken(c)
	if len(token) != 43 {
		return "", a.Unauthorized
	}
	var id string
	err := h.s.pool.QueryRow(ctx, `SELECT customer_id::text FROM customer_sessions WHERE token_hash=$1 AND expires_at>clock_timestamp()`, Hash(token)).Scan(&id)
	if err == pgx.ErrNoRows {
		return "", a.Unauthorized
	}
	return id, err
}
func (h *Handler) sessions(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()
	id, err := h.user(ctx, c)
	if err != nil {
		return err
	}
	rows, err := h.s.pool.Query(ctx, `SELECT token_hash=$2,expires_at FROM customer_sessions WHERE customer_id=$1 AND expires_at>now() ORDER BY expires_at DESC LIMIT 100`, id, Hash(currentToken(c)))
	if err != nil {
		return err
	}
	defer rows.Close()
	data := []fiber.Map{}
	for rows.Next() {
		var current bool
		var expires time.Time
		if err := rows.Scan(&current, &expires); err != nil {
			return err
		}
		data = append(data, fiber.Map{"current": current, "expiresAt": expires})
	}
	if err = rows.Err(); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": data})
}
func (h *Handler) revoke(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()
	id, err := h.user(ctx, c)
	if err != nil {
		return err
	}
	var in struct{}
	if err = decode(c, &in); err != nil {
		return err
	}
	tx, err := h.s.pool.Begin(ctx)
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)
	if _, err = tx.Exec(ctx, `DELETE FROM customer_sessions WHERE customer_id=$1 AND token_hash<>$2`, id, Hash(currentToken(c))); err != nil {
		return err
	}
	if _, err = tx.Exec(ctx, `INSERT INTO security_events(customer_id,event)VALUES($1,'other_sessions_revoked')`, id); err != nil {
		return err
	}
	if err = tx.Commit(ctx); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (s *Service) delivery(email, purpose string) (*a.AccountDelivery, error) {
	token := make([]byte, 32)
	if _, err := rand.Read(token); err != nil {
		return nil, err
	}
	raw := base64.RawURLEncoding.EncodeToString(token)
	duration := time.Hour
	subject := "Confirme seu e-mail na Forma"
	path := "confirmar-email"
	intro := "Confirme seu endereço de e-mail para concluir seu cadastro."
	if purpose == "reset" {
		duration = 30 * time.Minute
		subject = "Redefina sua senha na Forma"
		path = "redefinir-senha"
		intro = "Recebemos uma solicitação para redefinir sua senha. Se não foi você, ignore este e-mail."
	}
	body, err := encodeBody(s.key, intro+"\n\n"+s.origin+"/conta/"+path+"#token="+raw+"\n\nEste link tem validade limitada e pode ser utilizado uma vez.")
	if err != nil {
		return nil, err
	}
	return &a.AccountDelivery{TokenHash: Hash(raw), Email: email, Subject: subject, Body: body, ExpiresAt: time.Now().Add(duration)}, nil
}
