package app

import (
	"github.com/damonto/estkme-rlpa-server/internal/app/handler"
	"github.com/damonto/estkme-rlpa-server/internal/app/middleware"
	"github.com/damonto/estkme-rlpa-server/internal/pkg/rlpa"
	"github.com/gofiber/fiber/v3"
	"github.com/gofiber/fiber/v3/middleware/csrf"
	"github.com/gofiber/fiber/v3/middleware/recover"
	"github.com/gofiber/fiber/v3/middleware/requestid"
)

type Application interface {
	Run(address string) error
	Shutdown() error
}

type app struct {
	connManager rlpa.Manager
	fiber       *fiber.App
}

func New(connManager rlpa.Manager) Application {
	return &app{connManager: connManager}
}

func (a *app) Run(address string) error {
	a.fiber = fiber.New()
	a.fiber.Use(csrf.New())
	a.fiber.Use(requestid.New())
	a.fiber.Use(recover.New())
	a.fiber.Use(middleware.WithLpac(a.connManager))
	a.registerRoutes()
	return a.fiber.Listen(address)
}

func (a *app) registerRoutes() {
	api := a.fiber.Group("/api")
	{
		h := handler.NewChipHandler()
		api.Get("/chip", h.Info)
	}
}

func (a *app) Shutdown() error {
	return a.fiber.Shutdown()
}
