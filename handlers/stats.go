package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/minq3010/Backend-React-Native-App/db"
	"github.com/minq3010/Backend-React-Native-App/repositories"
)

func GetStatisticsHandler(ctx *fiber.Ctx) error {
	stats, err := repositories.GetStatistics(db.GetDB())
	if err != nil {
		return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{"error": err.Error()})
	}
	return ctx.Status(fiber.StatusOK).JSON(stats)
}