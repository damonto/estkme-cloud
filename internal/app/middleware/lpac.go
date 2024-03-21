package middleware

import (
	"github.com/damonto/estkme-rlpa-server/internal/pkg/lpac"
	"github.com/damonto/estkme-rlpa-server/internal/pkg/rlpa"
	"github.com/gofiber/fiber/v3"
)

type Token struct {
	Manager rlpa.Manager
}

func WithLpac(manager rlpa.Manager) fiber.Handler {
	return func(c fiber.Ctx) error {
		token := c.Query("token", c.Get("Authorization"))
		if token == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized, missing token",
			})
		}

		ctx := c.Context()
		conn, err := manager.Get(token)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized, invalid token",
			})
		}
		ctx.SetUserValue("lpac", lpac.NewCmder(conn.APDU))
		return c.Next()
	}
}
