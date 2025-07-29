package main

import (
	"fmt"

	"github.com/gofiber/fiber/v2"
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
	// repositories
	eventRepository := repositories.NewEventRepository(db)
	ticketRepository := repositories.NewTicketRepository(db)
	userRepository := repositories.NewUserRepository(db)
	authRepository := repositories.NewAuthRepository(db)
	paymentRepository := repositories.NewPaymentRepository(db)
	// service
	authService := services.NewAuthService(authRepository)

	// routing
	server := app.Group("/api")
	handlers.NewAuthHandler(server.Group("/auth"), authService)
	privateRoutes := server.Use(middlewares.AuthProtected(db))

	// handler
	handlers.NewEventHandler(privateRoutes.Group("/event"), eventRepository)
	handlers.NewTicketHandler(privateRoutes.Group("/ticket"), ticketRepository)
	handlers.NewUserHandler(privateRoutes.Group("/user"), userRepository)
	handlers.NewPaymentHandler(
		privateRoutes.Group("/payment"),
		paymentRepository,
		eventRepository,
		ticketRepository,
	)

	// for manager only
	server.Get("/stats", middlewares.ManagerOnly(), handlers.GetStatisticsHandler)

	// port
	app.Listen(fmt.Sprint(":" + envConfig.DBPort))
}
