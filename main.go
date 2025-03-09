package main

import (
	"trading-journal/config"
	"trading-journal/routes"

	"github.com/gofiber/fiber/v2"
)

func main() {
	app := fiber.New()
	// setup database
	config.ConnectDB()
	// register HTTP routes
	routes.SetupRoutes(app)
}
