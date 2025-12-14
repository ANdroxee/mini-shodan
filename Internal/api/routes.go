package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func (a *API) SetupRoutes() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Shodan@Home API v1.0",
		ServerHeader: "Shodan@Home",
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// Health check
	app.Get("/health", a.Health)

	// API v1
	v1 := app.Group("/api/v1")

	// Devices
	v1.Get("/devices", a.GetDevices)
	v1.Get("/devices/:ip", a.GetDeviceByIP)
	v1.Get("/devices/:ip/ports", a.GetPortsForDevice)

	// Stats
	v1.Get("/stats", a.GetStats)

	// Search
	v1.Get("/search", a.Search)

	// Alerts
	v1.Get("/alerts", a.GetAlerts)

	return app
}
