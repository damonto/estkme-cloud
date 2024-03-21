package handler

import "github.com/gofiber/fiber/v3"

type ChipHandler struct {
	Handler
}

func NewChipHandler() *ChipHandler {
	return &ChipHandler{}
}

type ChipInfo struct {
	EID       string  `json:"eid"`
	FreeSpace float32 `json:"freeSpace"`
}

func (c *ChipHandler) Info(ctx fiber.Ctx) error {
	chip, err := c.LpacCmder(ctx).Info()
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": err.Error(),
		})
	}

	return ctx.JSON(fiber.Map{
		"data": ChipInfo{
			EID:       chip.EID,
			FreeSpace: float32(chip.EUICCInfo2.ExtCardResource.FreeNonVolatileMemory) / 1024,
		},
	})
}
