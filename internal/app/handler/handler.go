package handler

import (
	"github.com/damonto/estkme-rlpa-server/internal/pkg/lpac"
	"github.com/gofiber/fiber/v3"
)

type Handler struct{}

func (Handler) LpacCmder(ctx fiber.Ctx) *lpac.Cmder {
	return ctx.Context().UserValue("lpac").(*lpac.Cmder)
}
