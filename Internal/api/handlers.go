package api

import (
	"shodan-home/internal/database"

	"github.com/gofiber/fiber/v2"
)

type API struct {
	db *database.DB
}

func NewAPI(db *database.DB) *API {
	return &API{db: db}
}

// Health check
func (a *API) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":    "ok",
		"timestamp": c.Context().Time(),
	})
}

// GetDevices retourne tous les devices
func (a *API) GetDevices(c *fiber.Ctx) error {
	filters := make(map[string]interface{})

	if service := c.Query("service"); service != "" {
		filters["service"] = service
	}

	if ip := c.Query("ip"); ip != "" {
		filters["ip"] = ip
	}

	if c.QueryBool("active") {
		filters["active"] = true
	}

	limit := c.QueryInt("limit", 100)
	filters["limit"] = limit

	devices, err := a.db.GetDevices(filters)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to fetch devices",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"devices": devices,
		"count":   len(devices),
	})
}

// GetDeviceByIP retourne un device spécifique
func (a *API) GetDeviceByIP(c *fiber.Ctx) error {
	ip := c.Params("ip")

	device, err := a.db.GetDeviceByIP(ip)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Device not found",
			"ip":    ip,
		})
	}

	return c.JSON(device)
}

// GetStats retourne les statistiques
func (a *API) GetStats(c *fiber.Ctx) error {
	stats, err := a.db.GetStats()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to fetch stats",
			"details": err.Error(),
		})
	}

	return c.JSON(stats)
}

// Search recherche dans devices et ports
func (a *API) Search(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return c.Status(400).JSON(fiber.Map{
			"error": "Query parameter 'q' is required",
		})
	}

	filters := map[string]interface{}{
		"ip":    query,
		"limit": 50,
	}

	devices, err := a.db.GetDevices(filters)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Search failed",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"results": devices,
		"query":   query,
		"count":   len(devices),
	})
}

// GetAlerts retourne les alertes non lues
func (a *API) GetAlerts(c *fiber.Ctx) error {
	alerts, err := a.db.GetUnreadAlerts()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error":   "Failed to fetch alerts",
			"details": err.Error(),
		})
	}

	return c.JSON(fiber.Map{
		"alerts": alerts,
		"count":  len(alerts),
	})
}

// GetPortsForDevice retourne les ports d'un device
func (a *API) GetPortsForDevice(c *fiber.Ctx) error {
	ip := c.Params("ip")

	device, err := a.db.GetDeviceByIP(ip)
	if err != nil {
		return c.Status(404).JSON(fiber.Map{
			"error": "Device not found",
		})
	}

	ports, err := a.db.GetPortsForDevice(device.ID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{
			"error": "Failed to fetch ports",
		})
	}

	return c.JSON(fiber.Map{
		"ip":    ip,
		"ports": ports,
		"count": len(ports),
	})
}
