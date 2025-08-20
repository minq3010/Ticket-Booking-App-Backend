package middlewares

import (
	"errors"
	"fmt"
	"os"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/golang-jwt/jwt"
	"github.com/minq3010/Backend-React-Native-App/models"
	"gorm.io/gorm"
)

func AuthProtected(db *gorm.DB) fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		path := ctx.Path()
		if path == "/api/payment-callback/momo-return" ||
			path == "/api/payment-callback/momo-ipn" {
			fmt.Printf("🔓 Skipping auth for callback: %s\n", path)
			return ctx.Next()
		}

		authHeader := ctx.Get("Authorization")
		if authHeader == "" {
			fmt.Printf("[Warn] empty authorization header for path: %s\n", path)
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"error": "missing authorization header",
			})
		}

		if authHeader == "" {
			log.Warn("empty athorization header")

			return ctx.Status(fiber.StatusUnauthorized).JSON(&fiber.Map{
				"status":  "fail",
				"message": "Unauthorized",
			})
		}
		tokenParts := strings.Split(authHeader, " ")

		if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
			log.Warn("invalid token parts")

			return ctx.Status(fiber.StatusUnauthorized).JSON(&fiber.Map{
				"status":  "fail",
				"message": "Unauthorized",
			})
		}
		tokenStr := tokenParts[1]
		secret := []byte(os.Getenv("JWT_SECRET"))

		token, err := jwt.Parse(tokenStr, func(token *jwt.Token) (interface{}, error) {
			if token.Method.Alg() != jwt.GetSigningMethod("HS256").Alg() {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return secret, nil
		})

		if err != nil || !token.Valid {
			log.Warn("invalid token")

			return ctx.Status(fiber.StatusUnauthorized).JSON(&fiber.Map{
				"status":  "fail",
				"message": "Unauthorized",
			})
		}

		userId := token.Claims.(jwt.MapClaims)["id"]
		if err := db.Model(&models.User{}).Where("id = ?", userId).Error; errors.Is(err, gorm.ErrRecordNotFound) {
			log.Warn("user not found in the db")

			return ctx.Status(fiber.StatusUnauthorized).JSON(&fiber.Map{
				"status":  "fail",
				"message": "Unauthorized",
			})
		}

		ctx.Locals("userId", userId)
		return ctx.Next()
	}
}

func ManagerOnly() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userId, ok := ctx.Locals("userId").(float64)
		if !ok {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "fail",
				"message": "User ID not found in context",
			})
		}

		if uint(userId) != 1 {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "fail",
				"message": "Access denied: You are not the Manager",
			})
		}

		return ctx.Next()
	}
}

func UserSelfOrManagerOnly() fiber.Handler {
	return func(ctx *fiber.Ctx) error {
		userIdFromToken, ok := ctx.Locals("userId").(float64)
		if !ok {
			return ctx.Status(fiber.StatusUnauthorized).JSON(fiber.Map{
				"status":  "fail",
				"message": "User ID not found in context",
			})
		}

		userIdFromParam := ctx.Params("userId")
		if userIdFromParam != fmt.Sprintf("%.0f", userIdFromToken) && uint(userIdFromToken) != 1 {
			return ctx.Status(fiber.StatusForbidden).JSON(fiber.Map{
				"status":  "fail",
				"message": "Access denied: You can only access your own info or you must be the Manager",
			})
		}
		return ctx.Next()
	}
}
