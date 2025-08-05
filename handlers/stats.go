package handlers

import (
	"context"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minq3010/Backend-React-Native-App/models"
)

type StatHandler struct {
	repository models.StatRepository
}

func (h *StatHandler) GetMany(ctx *fiber.Ctx) error {
	stats, err := h.repository.GetMany(ctx.Context())
	if err != nil {
		return ctx.Status(500).JSON(fiber.Map{
			"error": err.Error(),
		})	
	}
	return ctx.Status(fiber.StatusOK).JSON(stats)
}

func (h *StatHandler) GetOne(ctx *fiber.Ctx) error {
	eventId, _ := strconv.Atoi(ctx.Params("eventId"))

	context, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()
	stat, err := h.repository.GetOne(context, uint(eventId))

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"status": "fail",
			"error": err.Error(),
		})
	}
	
	return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
		"status": "success",
		"data": stat,
	})
}

func NewStatHandler(router fiber.Router, repository models.StatRepository) {
	handler := &StatHandler{
		repository: repository,
	}

	router.Get("/", handler.GetMany)
	router.Get("/:eventId", handler.GetOne)
}