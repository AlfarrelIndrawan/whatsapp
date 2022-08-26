package main

import (
	"log"

	"github.com/alfarrelindrawan/whatsapp/config"
	"github.com/alfarrelindrawan/whatsapp/handler"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	_ "github.com/mattn/go-sqlite3"
)

func setupRoutes(app *fiber.App) {
	app.Get("/", handler.Test)
	app.Post("/register", handler.Register)
	app.Post("/verify-otp", handler.VerifyOTP)
	app.Post("/login", handler.Login)
	app.Post("/resend-otp", handler.ResendOTP)
}

func main() {
	app := fiber.New()
	app.Use(cors.New())

	handler.WAConnect()
	config.Connect()

	setupRoutes(app)

	log.Fatal(app.Listen(":3000"))
}
