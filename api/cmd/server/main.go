package main

import (
	"forma/api/internal/accounts"
	cataloghttp "forma/api/internal/catalog/adapters/http"
	"forma/api/internal/catalog/adapters/postgres"
	"forma/api/internal/catalog/application"
	commercehttp "forma/api/internal/commerce/adapters/http"
	commercepg "forma/api/internal/commerce/adapters/postgres"
	"forma/api/internal/commerce/adapters/simulated"
	commerce "forma/api/internal/commerce/application"
	"forma/api/internal/contact"
	"forma/api/internal/platform"
	"github.com/gofiber/fiber/v3"
	"go.uber.org/fx"
	"go.uber.org/fx/fxevent"
	"go.uber.org/zap"
)

func main() {
	fx.New(
		fx.Provide(accounts.NewMailConfig, accounts.NewService, accounts.NewHandler, platform.NewConfig, platform.NewLogger, platform.NewDatabase, postgres.NewFactory, application.NewService, cataloghttp.NewHandler, platform.NewServer, commercepg.NewFactory, commerce.NewService, simulated.NewGateway, simulated.NewShipping, commercehttp.NewHandler, contact.NewRepository, contact.NewHandler),
		fx.WithLogger(func(log *zap.Logger) fxevent.Logger { return &fxevent.ZapLogger{Logger: log} }),
		fx.Invoke(accounts.RunMailer, func(app *fiber.App, h *accounts.Handler) { h.Register(app) }, func(s *commerce.Service, account *accounts.Service) { s.SetRegistrationPreparer(account.Prepare) }, commercepg.RunMaintenance, func(app *fiber.App, h *contact.Handler) { h.Register(app) }, func(app *fiber.App, handler *commercehttp.Handler) { handler.Register(app) }, platform.RunServer),
	).Run()
}
