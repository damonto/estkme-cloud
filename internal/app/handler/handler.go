package handler

import (
	"github.com/damonto/estkme-rlpa-server/internal/pkg/lpac"
	"github.com/damonto/estkme-rlpa-server/internal/pkg/rlpa"
	"github.com/gofiber/fiber/v3"
)

type Handler struct{}

func (h Handler) GetRLPAConn(ctx fiber.Ctx) *rlpa.Connection {
	return ctx.Context().UserValue("conn").(*rlpa.Connection)
}

func (h Handler) LpacCmder(ctx fiber.Ctx) *lpac.Cmder {
	conn := h.GetRLPAConn(ctx)
	return lpac.NewCmder(conn.APDU)
}
