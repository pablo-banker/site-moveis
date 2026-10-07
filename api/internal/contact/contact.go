package contact

import (
	"bytes"
	"context"
	"encoding/json"
	a "forma/api/internal/commerce/application"
	"github.com/gofiber/fiber/v3"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"net/mail"
	"strings"
	"time"
)

type Message struct {
	Name    string `json:"name"`
	Email   string `json:"email"`
	Subject string `json:"subject"`
	Body    string `json:"body"`
}
type Repository interface {
	Save(context.Context, string, Message) error
}
type postgres struct{ pool *pgxpool.Pool }

func NewRepository(pool *pgxpool.Pool) Repository { return &postgres{pool} }
func (r *postgres) Save(ctx context.Context, id string, m Message) error {
	_, err := r.pool.Exec(ctx, `INSERT INTO contact_messages(id,name,email,subject,body)VALUES($1,$2,$3,$4,$5)`, id, m.Name, m.Email, m.Subject, m.Body)
	return err
}

type Handler struct{ repo Repository }

func NewHandler(repo Repository) *Handler { return &Handler{repo} }
func (h *Handler) Register(app *fiber.App) {
	app.Post("/api/v1/contact", h.create)
}
func (h *Handler) create(c fiber.Ctx) error {
	if len(c.Body()) > 8192 {
		return a.Invalid("Mensagem muito grande.")
	}
	var m Message
	d := json.NewDecoder(bytes.NewReader(c.Body()))
	d.DisallowUnknownFields()
	if d.Decode(&m) != nil || d.Decode(new(any)) != io.EOF {
		return a.Invalid("Confira os dados da mensagem.")
	}
	m.Name = strings.TrimSpace(m.Name)
	m.Email = strings.TrimSpace(m.Email)
	m.Subject = strings.TrimSpace(m.Subject)
	m.Body = strings.TrimSpace(m.Body)
	email, err := mail.ParseAddress(m.Email)
	if err != nil || email.Address != m.Email || len(m.Email) > 254 || len(m.Name) < 2 || len(m.Name) > 100 || len(m.Subject) < 2 || len(m.Subject) > 100 || len(m.Body) < 5 || len(m.Body) > 5000 {
		return a.Invalid("Informe nome, e-mail, assunto e mensagem válidos.")
	}
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()
	id := uuid.NewString()
	if err := h.repo.Save(ctx, id, m); err != nil {
		return err
	}
	return c.Status(201).JSON(fiber.Map{"id": id})
}
