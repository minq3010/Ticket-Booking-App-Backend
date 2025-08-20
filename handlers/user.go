package handlers

import (
	"context"
	"os"
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/minq3010/Backend-React-Native-App/middlewares"
	"github.com/minq3010/Backend-React-Native-App/models"
	"github.com/minq3010/Backend-React-Native-App/utils"
)

type UserHandler struct {
	repository models.UserRepository
}

func (h *UserHandler) GetUserInfo(ctx *fiber.Ctx) error {
	userId, _ := strconv.Atoi(ctx.Params("userId"))

	context, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
	defer cancel()

	user, err := h.repository.GetUserInfo(context, uint(userId))

	if err != nil {
		return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
			"status":  "fail",
			"message": err.Error(),
		})
	}

	return ctx.Status(fiber.StatusOK).JSON(&fiber.Map{
		"status":  "success",
		"message": "this is your information",
		"data":    user,
	})
}

func (h *UserHandler) GetAllUsers(ctx *fiber.Ctx) error {
    email := ctx.Query("email")

    context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    var users []*models.User
    var err error

    if email != "" {
        users, err = h.repository.SearchUserAccountByEmail(context, email)
    } else {
        users, err = h.repository.GetAllUsers(context)
    }
    
    if err != nil {
        return ctx.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
            "status": "fail",
            "message": err.Error(),
        })
    }

    return ctx.Status(fiber.StatusOK).JSON(&fiber.Map{
        "status": "success",
        "data": users,
    })
}

func (h *UserHandler) UpdateUserInfo(ctx *fiber.Ctx) error {
    userId, _ := strconv.Atoi(ctx.Params("userId"))
    updateData := make(map[string]interface{})

    context, cancel := context.WithTimeout(context.Background(), time.Duration(5*time.Second))
    defer cancel()

    form, err := ctx.MultipartForm()
    if err != nil {
        return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
            "status": "fail",
            "message": "Cannot parse multipart form: " + err.Error(),
            "data": nil,
        })
    }

    if names, ok := form.Value["name"]; ok && len(names) > 0 {
        updateData["name"] = names[0]
    }
    if phones, ok := form.Value["phone"]; ok && len(phones) > 0 {
        updateData["phone"] = phones[0]
    }

    if files, ok := form.File["avatar"]; ok && len(files) > 0 {
        file := files[0]
        if err := os.MkdirAll("./tmp", os.ModePerm); err != nil {
            return ctx.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
                "status": "fail",
                "message": "Cannot create tmp folder: " + err.Error(),
                "data": nil,
            })
        }
        savePath := "./tmp/" + file.Filename
        if err := ctx.SaveFile(file, savePath); err != nil {
            return ctx.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
                "status": "fail",
                "message": "Cannot save file: " + err.Error(),
                "data": nil,
            })
        }
        avatarUrl, err := utils.UploadImageToCloudinary(savePath)
        os.Remove(savePath)
        if err != nil {
            return ctx.Status(fiber.StatusInternalServerError).JSON(&fiber.Map{
                "status": "fail",
                "message": "Cannot upload avatar: " + err.Error(),
                "data": nil,
            })
        }
        updateData["avatar"] = avatarUrl
    }

    user, err := h.repository.UpdateUserInfo(context, uint(userId), updateData)
    if err != nil {
        return ctx.Status(fiber.StatusBadRequest).JSON(&fiber.Map{
            "status": "fail",
            "message": err.Error(),
            "data": nil,
        })
    }

    return ctx.Status(fiber.StatusCreated).JSON(&fiber.Map{
        "status": "success",
        "message": "user updated",
        "data": user,
    })
}

func (h *UserHandler) DeleteUser(ctx *fiber.Ctx) error {
    userId, _ := strconv.Atoi(ctx.Params("userId"))
    context, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    err := h.repository.DeleteUser(context, uint(userId))

    if err != nil {
        return ctx.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
            "status": "fail",
            "message": err.Error,
        })
    }

    return ctx.Status(fiber.StatusOK).JSON(fiber.Map{
        "status": "success",
        "message": "User deleted",
    })
}

func NewUserHandler(router fiber.Router, repository models.UserRepository) {
	handler := &UserHandler{
		repository: repository,
	}

	router.Get("/:userId", middlewares.UserSelfOrManagerOnly(), handler.GetUserInfo)
	router.Put("/:userId", middlewares.UserSelfOrManagerOnly(), handler.UpdateUserInfo)
    router.Get("/", middlewares.ManagerOnly(), handler.GetAllUsers)
    router.Delete("/:userId", middlewares.ManagerOnly(), handler.DeleteUser)
}