package middleware

import (
	"strings"

	"github.com/damonto/estkme-rlpa-server/internal/pkg/rlpa"
	"github.com/gofiber/fiber/v3"
)

type Token struct {
	Manager rlpa.Manager
}

func WithRLPAConn(manager rlpa.Manager) fiber.Handler {
	return func(c fiber.Ctx) error {
		pin := strings.Replace(c.Query("pin_code", c.Get("Authorization")), "Bearer ", "", 1)
		if pin == "" {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized, missing pin code",
			})
		}

		ctx := c.Context()
		conn, err := manager.Get(pin)
		if err != nil {
			return c.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "unauthorized, invalid token",
			})
		}
		ctx.SetUserValue("conn", conn)
		return c.Next()
	}
}
