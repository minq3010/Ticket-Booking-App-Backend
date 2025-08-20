package handlers

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minq3010/Backend-React-Native-App/middlewares"
	"github.com/minq3010/Backend-React-Native-App/models"
	"github.com/skip2/go-qrcode"
)

type TicketHandler struct {
	repository models.TicketRepository
	statRepo 	models.StatRepository
}

func (h *TicketHandler) GetMany(ctx *fiber.Ctx) error {
	context, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()

	userId := uint(ctx.Locals("userId").(float64))

	tickets, err := h.repository.GetMany(context, userId)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"status":  "fail",
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(&fiber.Map{
		"status":  "success",
		"message": "",
		"data":    tickets,
	})
}

func (h *TicketHandler) GetTicketsByUser(ctx *fiber.Ctx) error {
    context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    userIdParam := ctx.Params("userId")
    userId, err := strconv.Atoi(userIdParam)
    if err != nil {
        return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
            "status": "fail",
            "message": "Invalid userId",
        })
    }

    tickets, err := h.repository.GetMany(context, uint(userId))
    if err != nil {
        return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
            "status": "fail",
            "message": err.Error(),
        })
    }

    return ctx.Status(fiber.StatusOK).JSON(&fiber.Map{
        "status": "success",
        "message": "",
        "data": tickets,
    })
}

func (h *TicketHandler) GetOne(ctx *fiber.Ctx) error {
	context, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()

	ticketId, _ := strconv.Atoi(ctx.Params("ticketId"))
	userId := uint(ctx.Locals("userId").(float64))

	ticket, err := h.repository.GetOne(context, userId, uint(ticketId))

	if err != nil {
		return ctx.Status(fiber.StatusBadGateway).JSON(&fiber.Map{
			"status":  "fail",
			"message": err.Error(),
		})
	}

	var QRCode []byte
	QRCode, err = qrcode.Encode (
		fmt.Sprintf("ticketId:%v,ownerId:%v", ticketId, userId),
		qrcode.Medium,
		256,
	)
	
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"status":  "fail",
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(&fiber.Map{
		"status":  "success",
		"message": "",
		"data": &fiber.Map{
			"ticket": ticket,
			"qrcode": QRCode,
		},
	})
}

func (h *TicketHandler) CreateOne(ctx *fiber.Ctx) error {
	context, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()

	ticket := &models.Ticket{}
	userId := uint(ctx.Locals("userId").(float64))

	if err := ctx.BodyParser(ticket); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(&fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data": nil,
			})	
	}

	ticket, err := h.repository.CreateOne(context, userId, ticket)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"status":  "fail",
			"message": err.Error(),
		})
	}


	return ctx.Status(fiber.StatusCreated).JSON(&fiber.Map{
		"status":  "success",
		"message": "",
		"data":    ticket,
	})	
}

func (h *TicketHandler) ValidateOne(ctx *fiber.Ctx) error {
	context, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()

	validateBody := &models.ValidateTicket{}

	if err := ctx.BodyParser(validateBody); err != nil {
		return ctx.Status(fiber.StatusUnprocessableEntity).JSON(&fiber.Map{
				"status":  "fail",
				"message": err.Error(),
				"data": nil,
			})	
	}

	validateData := make(map[string]interface{})
	validateData["entered"] = true

	ticket, err := h.repository.UpdateOne(context, validateBody.OwnerId, validateBody.TicketId, validateData)

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"status":  "fail",
			"message": err.Error(),
		})
	}
	
	go h.statRepo.UpdateStat(ctx.Context(), ticket.EventID)

	return ctx.Status(fiber.StatusOK).JSON(&fiber.Map{
		"status":  "success",
		"message": "Welcome to the show!",
		"data":    ticket,
	})	
}

func (h *TicketHandler) DeleteOne(ctx *fiber.Ctx) error {
	ticketId, _ := strconv.Atoi(ctx.Params("ticketId"))
	userId := uint(ctx.Locals("userId").(float64))

	context, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()

	ticket, err := h.repository.GetOne(context, userId, uint(ticketId))
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"status": "fail",
			"message": "Ticket not found",
		})
	}

	err = h.repository.DeleteOne(context, uint(ticketId))

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"status":  "fail",
			"message": err.Error(),
		})
	}

	go h.statRepo.UpdateStat(ctx.Context(), ticket.EventID)
	return ctx.SendStatus(fiber.StatusNoContent)
}


func NewTicketHandler(router fiber.Router, ticketRepo models.TicketRepository, statRepo models.StatRepository) {
	handler := &TicketHandler{
		repository: ticketRepo,
		statRepo: statRepo,
	}

	router.Get("/", handler.GetMany)
	router.Post("/", handler.CreateOne)
	router.Get("/:ticketId", handler.GetOne)
	router.Post("/validate", handler.ValidateOne)
	router.Delete("/:ticketId", handler.DeleteOne)

	// for manager
	router.Get("user/:userId", middlewares.ManagerOnly(), handler.GetTicketsByUser)
}