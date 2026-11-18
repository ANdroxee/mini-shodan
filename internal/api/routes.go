package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
)

func (a *API) SetupRoutes() *fiber.App {
	app := fiber.New(fiber.Config{
		AppName:      "Shodan@Home API v1.1",
		ServerHeader: "Shodan@Home",
	})

	app.Use(recover.New())
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowMethods: "GET,POST,PATCH,OPTIONS",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	app.Get("/health", a.Health)

	v1 := app.Group("/api/v1")

	v1.Get("/summary", a.GetSummary)

	v1.Get("/devices", a.GetDevices)
	v1.Get("/devices/:ip", a.GetDeviceByIP)
	v1.Patch("/devices/:ip", a.UpdateDevice)
	v1.Get("/devices/:ip/ports", a.GetPortsForDevice)

	v1.Get("/alerts", a.GetAlerts)
	v1.Post("/alerts/read-all", a.MarkAllRead)
	v1.Post("/alerts/:id/read", a.SetAlertRead)

	v1.Post("/scans", a.RequestScan)
	v1.Get("/scans/latest", a.GetLatestScan)

	v1.Get("/stats", a.GetStats)
	v1.Get("/search", a.Search)

	return app
}
