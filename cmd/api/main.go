package main

import (
	"context"
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/minq3010/Backend-React-Native-App/config"
	"github.com/minq3010/Backend-React-Native-App/db"
	"github.com/minq3010/Backend-React-Native-App/handlers"
	"github.com/minq3010/Backend-React-Native-App/middlewares"
	"github.com/minq3010/Backend-React-Native-App/repositories"
	"github.com/minq3010/Backend-React-Native-App/services"
)

func main() {
	envConfig := config.NewEnvConfig()
	db := db.Init(envConfig, db.DBMigrator)

	app := fiber.New(fiber.Config{
		AppName:      "TicketBooking",
		ServerHeader: "Fiber",
	})
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept, Authorization, ngrok-skip-browser-warning",
	}))

	// repositories
	eventRepository := repositories.NewEventRepository(db)
	ticketRepository := repositories.NewTicketRepository(db)
	userRepository := repositories.NewUserRepository(db)
	authRepository := repositories.NewAuthRepository(db)
	paymentRepository := repositories.NewPaymentRepository(db)
	statRepository := repositories.NewStatRepository(db)

	// kiểm tra và xoá sự kiện đã xảy ra > 2 ngày
	if err := eventRepository.DeleteExpiredEvents(context.Background()); err != nil {
		log.Printf("Error While Delete Expired Events: %v", err)
	}
	// update thống kê
	if err := statRepository.UpdateAllStats(context.Background()); err != nil {
		log.Printf("Error While Update Stats: %v", err)
	}

	// service
	authService := services.NewAuthService(authRepository)

	// routing
	server := app.Group("/api")
	handlers.NewAuthHandler(server.Group("/auth"), authService)
	privateRoutes := server.Use(middlewares.AuthProtected(db))

	// handler
	handlers.NewEventHandler(privateRoutes.Group("/event"), eventRepository, statRepository)
	handlers.NewTicketHandler(privateRoutes.Group("/ticket"), ticketRepository, statRepository)
	handlers.NewPaymentHandler(
		privateRoutes.Group("/payment"),
		paymentRepository,
		eventRepository,
		ticketRepository,
	)
	handlers.NewPaymentCallbackHandler(
		server.Group("/payment-callback"),
		paymentRepository,
		eventRepository,
		ticketRepository,
		statRepository,
	)
	// for users
	handlers.NewUserHandler(privateRoutes.Group("/user"), userRepository)
	// for manager only
	handlers.NewStatHandler(privateRoutes.Group("/manager/stat", middlewares.ManagerOnly()), statRepository)

	// port
	app.Listen(fmt.Sprint(":" + envConfig.DBPort))
}
