package platform

import (
	"context"
	"errors"
	"forma/api/docs"
	cataloghttp "forma/api/internal/catalog/adapters/http"
	"forma/api/internal/catalog/application"
	"forma/api/internal/catalog/domain"
	commerce "forma/api/internal/commerce/application"
	"forma/api/internal/traffic"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/cors"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"
	"go.uber.org/zap"
	"net"
	"strings"
	"time"
)

func NewServer(c Config, log *zap.Logger, pool *pgxpool.Pool, handler *cataloghttp.Handler) *fiber.App {
	app := fiber.New(fiber.Config{AppName: "Forma API", ReadTimeout: 10 * time.Second, WriteTimeout: 10 * time.Second, BodyLimit: 32768, ErrorHandler: func(ctx fiber.Ctx, err error) error {
		status, code, message := 500, "internal_error", "Não foi possível processar a solicitação."
		switch {
		case isCommerceError(err):
			var ce *commerce.Failure
			errors.As(err, &ce)
			status, code, message = ce.Status, ce.Code, ce.Message
		case errors.Is(err, domain.ErrNotFound):
			status, code, message = 404, "not_found", "Móvel não encontrado."
		case errors.Is(err, application.ErrInvalidFilter):
			status, code, message = 400, "invalid_filter", "Filtros inválidos."
		default:
			var fe *fiber.Error
			if errors.As(err, &fe) {
				status = fe.Code
				if status < 500 {
					code, message = "request_error", fe.Message
				}
			}
		}
		if status >= 500 {
			log.Error("request failed", zap.String("request_id", ctx.GetRespHeader("X-Request-ID")), zap.Error(err))
		}
		return ctx.Status(status).JSON(fiber.Map{"error": fiber.Map{"code": code, "message": message, "requestId": ctx.GetRespHeader("X-Request-ID")}})
	}})
	metrics := &metrics{}
	app.Use(metrics.observe)
	app.Get("/internal/metrics", metrics.handler(pool, c.MetricsToken))
	app.Use(func(ctx fiber.Ctx) error {
		ctx.Set("X-Content-Type-Options", "nosniff")
		ctx.Set("X-Frame-Options", "DENY")
		ctx.Set("Referrer-Policy", "no-referrer")
		ctx.Set("Content-Security-Policy", "default-src 'none'; frame-ancestors 'none'")
		return ctx.Next()
	})
	app.Use(requestid.New())
	app.Use(func(ctx fiber.Ctx) error {
		start := time.Now()
		err := ctx.Next()
		// Paths only: queries, bodies, cookies and authorization are not logged.
		log.Info("http request", zap.String("method", ctx.Method()), zap.String("path", ctx.Path()), zap.String("request_id", ctx.GetRespHeader("X-Request-ID")), zap.Duration("duration", time.Since(start)))
		return err
	})
	app.Use(recover.New())
	app.Use(cors.New(cors.Config{AllowOrigins: strings.Split(c.AllowedOrigins, ","), AllowMethods: []string{"GET", "POST", "OPTIONS"}, AllowHeaders: []string{"Content-Type", "Authorization"}}))
	app.Get("/health/live", func(ctx fiber.Ctx) error { return ctx.JSON(fiber.Map{"status": "ok"}) })
	app.Get("/health/ready", func(ctx fiber.Ctx) error {
		ping, cancel := context.WithTimeout(ctx.Context(), time.Second)
		defer cancel()
		if err := pool.Ping(ping); err != nil {
			return ctx.Status(503).JSON(fiber.Map{"status": "unavailable"})
		}
		return ctx.JSON(fiber.Map{"status": "ok"})
	})
	if c.EnableDocs {
		app.Get("/openapi.json", func(ctx fiber.Ctx) error { ctx.Set("Content-Type", "application/json"); return ctx.Send(docs.Spec) })
		app.Get("/docs", func(ctx fiber.Ctx) error {
			ctx.Set("Content-Security-Policy", "default-src 'none'; script-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; style-src 'self' 'unsafe-inline' https://cdn.jsdelivr.net; img-src 'self' data:; connect-src 'self'; frame-ancestors 'none'")
			ctx.Set("Content-Type", "text/html; charset=utf-8")
			return ctx.Send(docs.UI)
		})
	}
	app.Use("/api/v1", traffic.Middleware(pool, c.BFFSecret, c.Environment == "production"))
	app.Post("/internal/traffic", func(ctx fiber.Ctx) error {
		client, err := traffic.Client(ctx, c.BFFSecret, c.Environment == "production")
		if err != nil {
			return err
		}
		work, cancel := context.WithTimeout(ctx.Context(), 2*time.Second)
		defer cancel()
		allowed, err := traffic.Allow(work, pool, "web:"+client, 120, time.Minute)
		if err != nil {
			return err
		}
		if !allowed {
			ctx.Set("Retry-After", "60")
			return fiber.ErrTooManyRequests
		}
		return ctx.JSON(fiber.Map{"ok": true})
	})
	app.Get("/api/v1/addresses/:cep", traffic.AddressHandler(pool))
	handler.Register(app)
	return app
}
func RunServer(lc fx.Lifecycle, app *fiber.App, c Config, log *zap.Logger, shutdown fx.Shutdowner) {
	lc.Append(fx.Hook{OnStart: func(context.Context) error {
		listener, err := net.Listen("tcp", c.Address)
		if err != nil {
			return err
		}
		go func() {
			if err := app.Listener(listener, fiber.ListenConfig{DisableStartupMessage: true}); err != nil {
				log.Error("server stopped", zap.Error(err))
				_ = shutdown.Shutdown(fx.ExitCode(1))
			}
		}()
		log.Info("API listening", zap.String("address", c.Address))
		return nil
	}, OnStop: app.ShutdownWithContext})
}

func isCommerceError(err error) bool { var ce *commerce.Failure; return errors.As(err, &ce) }
