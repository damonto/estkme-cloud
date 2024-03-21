package handler

import (
	"github.com/damonto/estkme-rlpa-server/internal/pkg/rlpa"
	"github.com/gofiber/fiber/v3"
)

type ConnectHandler struct {
	Handler
	rlpaConnManager rlpa.Manager
}

func NewConnectHandler(manager rlpa.Manager) *ConnectHandler {
	return &ConnectHandler{
		rlpaConnManager: manager,
	}
}

type ConnectRequest struct {
	PinCode string `json:"pinCode"`
}

func (h *ConnectHandler) Connect(ctx fiber.Ctx) error {
	var req ConnectRequest
	if err := ctx.Bind().JSON(&req); err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	if _, err := h.rlpaConnManager.Get(req.PinCode); err != nil {
		return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
			"error": err.Error(),
		})
	}
	return ctx.JSON(fiber.Map{"data": "success"})
}
