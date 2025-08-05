package handlers

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minq3010/Backend-React-Native-App/models"
	"github.com/minq3010/Backend-React-Native-App/utils"
)

type EventHandler struct {
	repository models.EventRepository
}

func (h *EventHandler) GetMany(ctx *fiber.Ctx) error {
	name := ctx.Query("name")


	context, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()
	
	var events []*models.Event
	var err error
	if name != "" {
		events, err = h.repository.SearchByName(context, name)
	} else {
		events, err = h.repository.GetMany(context)
	}

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"status":  "fail",
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(&fiber.Map{
		"status":  "success",
		"message": "",
		"data":    events,
	})
}

func (h *EventHandler) GetOne(ctx *fiber.Ctx) error {
	eventId, _ := strconv.Atoi(ctx.Params("eventId"))

	context, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()

	event, err := h.repository.GetOne(context, uint(eventId))

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"status":  "fail",
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(&fiber.Map{
		"status":  "success",
		"message": "",
		"data":    event,
	})
}

func (h *EventHandler) CreateOne(ctx *fiber.Ctx) error {
	event := &models.Event{}

	context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// ✅ Parse multipart form để hỗ trợ file upload
	form, err := ctx.MultipartForm()
	if err != nil {
		// Fallback to JSON parsing if not multipart
		if err := ctx.BodyParser(event); err != nil {
			return ctx.Status(fiber.StatusUnprocessableEntity).JSON(&fiber.Map{
				"status":  "fail",
				"message": "Cannot parse request: " + err.Error(),
				"data":    nil,
			})
		}
	} else {
		// ✅ Lấy các trường text từ form-data
		if names, ok := form.Value["name"]; ok && len(names) > 0 {
			event.Name = names[0]
		}
		if locations, ok := form.Value["location"]; ok && len(locations) > 0 {
			event.Location = locations[0]
		}
		if descriptions, ok := form.Value["description"]; ok && len(descriptions) > 0 {
			event.Description = descriptions[0]
		}
		if prices, ok := form.Value["price"]; ok && len(prices) > 0 {
			if price, err := strconv.ParseInt(prices[0], 10, 64); err == nil {
				event.Price = price
			}
		}
		if maxTickets, ok := form.Value["maxTickets"]; ok && len(maxTickets) > 0 {
			if maxTickets, err := strconv.ParseInt(maxTickets[0], 10, 64); err == nil {
				event.MaxTickets = maxTickets
			}
		}
		if dates, ok := form.Value["date"]; ok && len(dates) > 0 {
			if date, err := time.Parse(time.RFC3339, dates[0]); err == nil {
				event.Date = date
			}
		}

		// ✅ Xử lý file image upload
		if files, ok := form.File["image"]; ok && len(files) > 0 {
			file := files[0]

			// Tạo thư mục tmp nếu chưa có
			if err := os.MkdirAll("./tmp", os.ModePerm); err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
					"status":  "fail",
					"message": "Cannot create tmp folder: " + err.Error(),
					"data":    nil,
				})
			}

			// Lưu file tạm thời
			savePath := "./tmp/" + file.Filename
			if err := ctx.SaveFile(file, savePath); err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
					"status":  "fail",
					"message": "Cannot save file: " + err.Error(),
					"data":    nil,
				})
			}

			// Upload lên Cloudinary
			imageUrl, err := utils.UploadImageToCloudinary(savePath)

			// Xóa file tạm
			os.Remove(savePath)

			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
					"status":  "fail",
					"message": "Cannot upload image: " + err.Error(),
					"data":    nil,
				})
			}

			event.ImageURL = imageUrl
		}
	}

	// ✅ Validate giá vé
	if event.Price < 0 {
		return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"status":  "fail",
			"message": "Price must be greater than or equal to 0",
		})
	}

	// ✅ Validate required fields
	if event.Name == "" {
		return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"status":  "fail",
			"message": "Event name is required",
		})
	}

	event, err = h.repository.CreateOne(context, event)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"status":  "fail",
			"message": err.Error(),
			"data":    nil,
		})
	}

	return ctx.Status(fiber.StatusCreated).JSON(&fiber.Map{
		"status":  "success",
		"message": "Event created",
		"data":    event,
	})
}

func (h *EventHandler) UpdateOne(ctx *fiber.Ctx) error {
	eventId, _ := strconv.Atoi(ctx.Params("eventId"))
	updateData := make(map[string]interface{})

	context, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()

	// Try to parse as multipart form first (like user handler)
	form, err := ctx.MultipartForm()
	if err != nil {
		// If multipart fails, try JSON parsing into a struct then convert to map
		event := &models.Event{}
		if err := ctx.BodyParser(event); err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
				"status":  "fail",
				"message": "Cannot parse request: " + err.Error(),
				"data":    nil,
			})
		}

		// Convert struct to map for repository
		if event.Name != "" {
			updateData["name"] = event.Name
		}
		if event.Description != "" {
			updateData["description"] = event.Description
		}
		if event.Location != "" {
			updateData["location"] = event.Location
		}
		if !event.Date.IsZero() {
			updateData["date"] = event.Date
		}
		if event.Price != 0 {
			updateData["price"] = event.Price
		}
		if event.MaxTickets != 0 {
			updateData["max_tickets"] = event.MaxTickets
		}
		if event.ImageURL != "" {
			updateData["image_url"] = event.ImageURL
		}
	} else {
		// Handle multipart form data (like user handler)
		if names, ok := form.Value["name"]; ok && len(names) > 0 {
			updateData["name"] = names[0]
		}
		if descriptions, ok := form.Value["description"]; ok && len(descriptions) > 0 {
			updateData["description"] = descriptions[0]
		}
		if locations, ok := form.Value["location"]; ok && len(locations) > 0 {
			updateData["location"] = locations[0]
		}
		if dates, ok := form.Value["date"]; ok && len(dates) > 0 {
			if date, err := time.Parse(time.RFC3339, dates[0]); err == nil {
				updateData["date"] = date
			}
		}
		if prices, ok := form.Value["price"]; ok && len(prices) > 0 {
			if price, err := strconv.ParseInt(prices[0], 10, 64); err == nil {
				updateData["price"] = price
			}
		}
		if maxTickets, ok := form.Value["maxTickets"]; ok && len(maxTickets) > 0 {
			if maxTickets, err := strconv.ParseInt(maxTickets[0], 10, 64); err == nil {
				updateData["max_tickets"] = maxTickets
			}
		}

		// Handle image upload (like user handler)
		if files, ok := form.File["image"]; ok && len(files) > 0 {
			file := files[0]
			if err := os.MkdirAll("./tmp", os.ModePerm); err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
					"status":  "fail",
					"message": "Cannot create tmp folder: " + err.Error(),
					"data":    nil,
				})
			}
			savePath := "./tmp/" + file.Filename
			if err := ctx.SaveFile(file, savePath); err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
					"status":  "fail",
					"message": "Cannot save file: " + err.Error(),
					"data":    nil,
				})
			}
			imageUrl, err := utils.UploadImageToCloudinary(savePath)
			os.Remove(savePath)
			if err != nil {
				return ctx.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
					"status":  "fail",
					"message": "Cannot upload image: " + err.Error(),
					"data":    nil,
				})
			}
			updateData["image_url"] = imageUrl
		}
	}

	// Validate price if provided
	if val, ok := updateData["price"]; ok {
		switch v := val.(type) {
		case float64:
			if v < 0 {
				return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
					"status":  "fail",
					"message": "Price must be >= 0",
				})
			}
		case int:
			if v < 0 {
				return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
					"status":  "fail",
					"message": "Price must be >= 0",
				})
			}
		}
	}

	event, err := h.repository.UpdateOne(context, uint(eventId), updateData)
	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"status":  "fail",
			"message": err.Error(),
			"data":    nil,
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(&fiber.Map{
		"status":  "success",
		"message": "Event updated",
		"data":    event,
	})
}

func (h *EventHandler) DeleteOne(ctx *fiber.Ctx) error {
	eventId, _ := strconv.Atoi(ctx.Params("eventId"))

	context, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()

	err := h.repository.DeleteOne(context, uint(eventId))

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"status":  "fail",
			"message": err.Error(),
		})
	}

	return ctx.SendStatus(fiber.StatusNoContent)
}

func NewEventHandler(router fiber.Router, repository models.EventRepository) {
	handler := &EventHandler{
		repository: repository,
	}

	router.Post("/", handler.CreateOne)
	router.Get("/", handler.GetMany)
	router.Get("/:eventId", handler.GetOne)
	router.Put("/:eventId", handler.UpdateOne)
	router.Delete("/:eventId", handler.DeleteOne)
}
