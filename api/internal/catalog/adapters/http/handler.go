package http

import (
	"context"
	"forma/api/internal/catalog/application"
	"forma/api/internal/catalog/domain"
	"github.com/gofiber/fiber/v3"
	"strconv"
	"time"
)

type Handler struct{ service *application.Service }

func NewHandler(service *application.Service) *Handler { return &Handler{service: service} }

type Product struct {
	ID         string   `json:"id"`
	Name       string   `json:"name"`
	Category   string   `json:"category"`
	Room       string   `json:"room"`
	Price      float64  `json:"price"`
	PriceCents int64    `json:"priceCents"`
	Image      string   `json:"image"`
	Dimensions string   `json:"dimensions"`
	Material   string   `json:"material"`
	Label      string   `json:"label,omitempty"`
	Finishes   []string `json:"finishes"`
}

func dto(p domain.Product) Product {
	return Product{p.ID, p.Name, p.Category, p.Room, float64(p.PriceCents) / 100, p.PriceCents, p.Image, p.Dimensions, p.Material, p.Label, p.Finishes}
}
func (h *Handler) Register(app *fiber.App) {
	app.Get("/api/v1/products", h.List)
	app.Get("/api/v1/products/:id", h.Find)
	app.Get("/api/v1/categories", h.Categories)
	app.Get("/api/v1/rooms", h.Rooms)
}
func param(c fiber.Ctx, name string, fallback int) (int, error) {
	raw := c.Query(name)
	if raw == "" {
		return fallback, nil
	}
	value, err := strconv.Atoi(raw)
	if err != nil {
		return 0, application.ErrInvalidFilter
	}
	return value, nil
}
func (h *Handler) List(c fiber.Ctx) error {
	limit, err := param(c, "limit", 24)
	if err != nil || limit < 1 {
		return application.ErrInvalidFilter
	}
	offset, err := param(c, "offset", 0)
	if err != nil {
		return err
	}
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()
	result, err := h.service.List(ctx, domain.Filter{Query: c.Query("q"), Category: c.Query("category"), Room: c.Query("room"), Sort: c.Query("sort"), Limit: limit, Offset: offset})
	if err != nil {
		return err
	}
	data := make([]Product, 0, len(result.Products))
	for _, p := range result.Products {
		data = append(data, dto(p))
	}
	return c.JSON(fiber.Map{"data": data, "total": result.Total, "limit": limit, "offset": offset})
}
func (h *Handler) Find(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()
	p, err := h.service.Find(ctx, c.Params("id"))
	if err != nil {
		return err
	}
	return c.JSON(dto(p))
}
func (h *Handler) Categories(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()
	data, err := h.service.Categories(ctx)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": data})
}
func (h *Handler) Rooms(c fiber.Ctx) error {
	ctx, cancel := context.WithTimeout(c.Context(), 3*time.Second)
	defer cancel()
	data, err := h.service.Rooms(ctx)
	if err != nil {
		return err
	}
	return c.JSON(fiber.Map{"data": data})
}
