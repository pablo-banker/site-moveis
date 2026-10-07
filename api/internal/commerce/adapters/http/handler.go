package http

import (
	"bytes"
	"context"
	"encoding/json"
	a "forma/api/internal/commerce/application"
	"forma/api/internal/traffic"
	"github.com/gofiber/fiber/v3"
	"github.com/jackc/pgx/v5/pgxpool"
	"io"
	"strings"
	"time"
)

type Handler struct {
	s    *a.Service
	pool *pgxpool.Pool
}

func NewHandler(s *a.Service, pool *pgxpool.Pool) *Handler { return &Handler{s: s, pool: pool} }
func decode(c fiber.Ctx, out any) error {
	if strings.Split(c.Get("Content-Type"), ";")[0] != "application/json" {
		return &a.Failure{Status: 415, Code: "unsupported_media_type", Message: "Envie os dados como JSON."}
	}
	if len(c.Body()) > 32768 {
		return a.Invalid("Solicitação muito grande.")
	}
	d := json.NewDecoder(bytes.NewReader(c.Body()))
	d.DisallowUnknownFields()
	if err := d.Decode(out); err != nil {
		return a.Invalid("Confira os dados enviados.")
	}
	if d.Decode(new(any)) != io.EOF {
		return a.Invalid("JSON inválido.")
	}
	return nil
}
func token(c fiber.Ctx) string {
	value := c.Get("Authorization")
	if !strings.HasPrefix(value, "Bearer ") {
		return ""
	}
	return strings.TrimPrefix(value, "Bearer ")
}
func (h *Handler) Register(app *fiber.App) {
	app.Use("/api/v1/auth", func(c fiber.Ctx) error { c.Set("Cache-Control", "no-store"); return c.Next() })
	app.Post("/api/v1/auth/register", h.register)
	app.Post("/api/v1/auth/login", h.login)
	app.Get("/api/v1/auth/me", h.me)
	app.Post("/api/v1/auth/profile", h.updateProfile)
	app.Get("/api/v1/products/:id/variants", h.variants)
	app.Post("/api/v1/auth/logout", h.logout)
	app.Use("/api/v1/orders", func(c fiber.Ctx) error { c.Set("Cache-Control", "no-store"); return c.Next() })
	app.Post("/api/v1/orders", h.create)
	app.Post("/api/v1/shipping/quotes", h.quoteShipping)
	app.Get("/api/v1/orders", h.list)
	app.Get("/api/v1/orders/:id", h.find)
	app.Post("/api/v1/orders/:id/cancel", h.cancel)
	app.Post("/api/v1/orders/:id/payment", h.pay)
}
func (h *Handler) register(c fiber.Ctx) error {
	var in a.AuthInput
	if err := decode(c, &in); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	_, err := h.s.Register(ctx, in)
	if err != nil && err != a.Conflict {
		return err
	}
	return c.Status(202).JSON(fiber.Map{"message": "Se o cadastro puder ser concluído, você receberá um e-mail para confirmar sua conta."})
}
func (h *Handler) login(c fiber.Ctx) error {
	var in a.AuthInput
	if err := decode(c, &in); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	result, err := h.s.Login(ctx, in)
	if err != nil {
		return err
	}
	return c.JSON(result)
}
func (h *Handler) me(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()
	user, err := h.s.Current(ctx, token(c))
	if err != nil {
		return err
	}
	return c.JSON(user)
}
func (h *Handler) logout(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()
	if err := h.s.Logout(ctx, token(c)); err != nil {
		return err
	}
	return c.JSON(fiber.Map{"ok": true})
}

func (h *Handler) updateProfile(c fiber.Ctx) error {
	var in struct {
		Name string `json:"name"`
	}
	if err := decode(c, &in); err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()
	user, err := h.s.UpdateName(ctx, token(c), in.Name)
	if err != nil {
		return err
	}
	return c.JSON(user)
}
func (h *Handler) create(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	user, err := h.s.Current(ctx, token(c))
	if err != nil {
		return err
	}
	if !user.EmailVerified {
		return &a.Failure{Status: 403, Code: "email_unverified", Message: "Confirme seu e-mail antes de finalizar o pedido."}
	}
	if err := h.quota(ctx, user.ID, "orders", 20); err != nil {
		return err
	}
	var in a.OrderInput
	if err := decode(c, &in); err != nil {
		return err
	}
	order, err := h.s.CreateOrder(ctx, user.ID, in)
	if err != nil {
		return err
	}
	return c.Status(201).JSON(order)
}
func (h *Handler) list(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	user, err := h.s.Current(ctx, token(c))
	if err != nil {
		return err
	}
	orders, err := h.s.Orders(ctx, user.ID)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": orders})
}
func (h *Handler) find(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()
	user, err := h.s.Current(ctx, token(c))
	if err != nil {
		return err
	}
	o, err := h.s.Order(ctx, user.ID, c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(o)
}
func (h *Handler) cancel(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	user, err := h.s.Current(ctx, token(c))
	if err != nil {
		return err
	}
	o, err := h.s.Cancel(ctx, user.ID, c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(o)
}

func (h *Handler) pay(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	user, err := h.s.Current(ctx, token(c))
	if err != nil {
		return err
	}
	var in struct{}
	if err := decode(c, &in); err != nil {
		return err
	}
	o, err := h.s.Pay(ctx, user.ID, c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(o)
}

func (h *Handler) variants(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()
	data, err := h.s.Variants(ctx, c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": data})
}

func (h *Handler) quoteShipping(c fiber.Ctx) error {
	c.Set("Cache-Control", "no-store")
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Second)
	defer cancel()
	user, err := h.s.Current(ctx, token(c))
	if err != nil {
		return err
	}
	if err := h.quota(ctx, user.ID, "shipping", 30); err != nil {
		return err
	}
	var in a.ShippingInput
	if err := decode(c, &in); err != nil {
		return err
	}
	quote, err := h.s.QuoteShipping(ctx, user.ID, in)
	if err != nil {
		return err
	}
	return c.JSON(quote)
}

func (h *Handler) quota(ctx context.Context, id, scope string, max int) error {
	allowed, err := traffic.Allow(ctx, h.pool, scope+":customer:"+id, max, time.Minute)
	if err != nil {
		return err
	}
	if !allowed {
		return &a.Failure{Status: 429, Code: "quota", Message: "Muitas solicitações. Aguarde um minuto."}
	}
	return nil
}
