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
	handlers.NewPaymentCallbackHandler(
		server.Group("/payment-callback"), // Không qua privateRoutes
		paymentRepository,
		eventRepository,
		ticketRepository,
	)
	 // ✅ Debug: Log tất cả routes
    fmt.Println("🔧 Registered routes:")
    fmt.Println("  POST /api/payment/momo (with auth)")
    fmt.Println("  GET  /api/payment-callback/momo-return (no auth)")
    fmt.Println("  POST /api/payment-callback/momo-ipn (no auth)")

	// for manager only
	server.Get("/stats", middlewares.ManagerOnly(), handlers.GetStatisticsHandler)

	// port
	app.Listen(fmt.Sprint(":" + envConfig.DBPort))
}
