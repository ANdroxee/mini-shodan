package api

import (
	"strconv"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/robfig/cron/v3"

	"shodan-home/internal/database"
	"shodan-home/internal/risk"
)

// Config : réglages du scan, affichés dans l'interface
type Config struct {
	Target   string
	Schedule string
	Timezone string
}

type API struct {
	db  *database.DB
	cfg Config
}

func NewAPI(db *database.DB, cfg Config) *API {
	return &API{db: db, cfg: cfg}
}

func fail(c *fiber.Ctx, code int, msg string, err error) error {
	body := fiber.Map{"error": msg}
	if err != nil {
		body["details"] = err.Error()
	}
	return c.Status(code).JSON(body)
}

// Health check
func (a *API) Health(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{"status": "ok", "timestamp": time.Now()})
}

// enrich calcule pour chaque appareil s'il est en ligne et son statut :
// off (pas vu au dernier scan), bad (port critique), warn (port risqué), new (nouveau, pas encore reconnu), ok
func (a *API) enrich(devices []database.Device) ([]database.Device, error) {
	last, err := a.db.LastCompletedScan(0)
	if err != nil {
		return nil, err
	}
	open, err := a.db.OpenAlertsByDevice()
	if err != nil {
		return nil, err
	}
	for i := range devices {
		d := &devices[i]
		if last != nil {
			d.Online = !d.LastSeen.Before(last.StartTime)
		} else {
			d.Online = time.Since(d.LastSeen) < 24*time.Hour
		}
		d.Status = "ok"
		isNew := false
		for _, al := range open[d.ID] {
			if al.Type == "new_device" {
				isNew = true
			}
		}
		worst := ""
		for _, p := range d.Ports {
			if p.Risk == risk.Critical || (p.Risk != "" && worst == "") {
				worst = p.Risk
			}
		}
		switch {
		case !d.Online:
			d.Status = "off"
		case worst == risk.Critical:
			d.Status = "bad"
		case worst != "":
			d.Status = "warn"
		case isNew && !d.Known:
			d.Status = "new"
		}
	}
	return devices, nil
}

// GetSummary : tout ce qu'il faut pour l'écran d'accueil
func (a *API) GetSummary(c *fiber.Ctx) error {
	devices, err := a.db.GetDevices(map[string]interface{}{})
	if err != nil {
		return fail(c, 500, "Failed to fetch devices", err)
	}
	if devices, err = a.enrich(devices); err != nil {
		return fail(c, 500, "Failed to compute status", err)
	}
	alerts, err := a.db.ListAlerts(false, 1000)
	if err != nil {
		return fail(c, 500, "Failed to fetch alerts", err)
	}
	lastScan, err := a.db.LastCompletedScan(0)
	if err != nil {
		return fail(c, 500, "Failed to fetch scans", err)
	}
	scanning, _ := a.db.ScanInProgress()

	score := 100
	online, offline := 0, 0
	for _, d := range devices {
		if !d.Online {
			offline++
			continue
		}
		online++
		for _, p := range d.Ports {
			score -= risk.Weight(p.Risk)
		}
		if d.Status == "new" {
			score -= 8
		}
	}
	if score < 0 {
		score = 0
	}

	issues := fiber.Map{"urgent": 0, "check": 0, "new": 0, "info": 0}
	unread := 0
	for _, al := range alerts {
		if !al.Read {
			unread++
		}
		switch {
		case al.Type == "risky_port" && al.Severity == risk.Critical:
			issues["urgent"] = issues["urgent"].(int) + 1
		case al.Type == "risky_port":
			issues["check"] = issues["check"].(int) + 1
		case al.Type == "new_device":
			issues["new"] = issues["new"].(int) + 1
		default:
			issues["info"] = issues["info"].(int) + 1
		}
	}

	var next *time.Time
	if sched, err := cron.ParseStandard(a.cfg.Schedule); err == nil {
		loc, err := time.LoadLocation(a.cfg.Timezone)
		if err != nil {
			loc = time.UTC
		}
		n := sched.Next(time.Now().In(loc)).UTC()
		next = &n
	}

	return c.JSON(fiber.Map{
		"score":         score,
		"devices":       fiber.Map{"total": len(devices), "online": online, "offline": offline},
		"issues":        issues,
		"unread_alerts": unread,
		"last_scan":     lastScan,
		"scanning":      scanning,
		"next_scan":     next,
		"target":        a.cfg.Target,
		"schedule":      a.cfg.Schedule,
	})
}

// GetDevices retourne les appareils (avec ports et statut)
func (a *API) GetDevices(c *fiber.Ctx) error {
	filters := map[string]interface{}{"limit": c.QueryInt("limit", 500)}
	if service := c.Query("service"); service != "" {
		filters["service"] = service
	}
	if ip := c.Query("ip"); ip != "" {
		filters["ip"] = ip
	}
	if c.QueryBool("active") {
		filters["active"] = true
	}

	devices, err := a.db.GetDevices(filters)
	if err != nil {
		return fail(c, 500, "Failed to fetch devices", err)
	}
	if devices, err = a.enrich(devices); err != nil {
		return fail(c, 500, "Failed to compute status", err)
	}
	return c.JSON(fiber.Map{"devices": devices, "count": len(devices)})
}

// GetDeviceByIP retourne un appareil
func (a *API) GetDeviceByIP(c *fiber.Ctx) error {
	device, err := a.db.GetDeviceByIP(c.Params("ip"))
	if err != nil {
		return fail(c, 404, "Device not found", nil)
	}
	list, err := a.enrich([]database.Device{*device})
	if err != nil {
		return fail(c, 500, "Failed to compute status", err)
	}
	return c.JSON(list[0])
}

// UpdateDevice : renommer un appareil, ou indiquer qu'on le connaît
func (a *API) UpdateDevice(c *fiber.Ctx) error {
	var body struct {
		Name  *string `json:"name"`
		Known *bool   `json:"known"`
	}
	if err := c.BodyParser(&body); err != nil {
		return fail(c, 400, "Invalid JSON body", err)
	}
	ip := c.Params("ip")
	if body.Name != nil {
		name := *body.Name
		if len([]rune(name)) > 100 {
			return fail(c, 400, "Name too long (100 characters max)", nil)
		}
		if err := a.db.SetDeviceName(ip, name); err != nil {
			return fail(c, 404, "Device not found", nil)
		}
	}
	if body.Known != nil {
		if err := a.db.SetDeviceKnown(ip, *body.Known); err != nil {
			return fail(c, 404, "Device not found", nil)
		}
	}
	return a.GetDeviceByIP(c)
}

// GetPortsForDevice retourne les ports d'un appareil
func (a *API) GetPortsForDevice(c *fiber.Ctx) error {
	ip := c.Params("ip")
	device, err := a.db.GetDeviceByIP(ip)
	if err != nil {
		return fail(c, 404, "Device not found", nil)
	}
	return c.JSON(fiber.Map{"ip": ip, "ports": device.Ports, "count": len(device.Ports)})
}

// GetStats retourne les statistiques
func (a *API) GetStats(c *fiber.Ctx) error {
	stats, err := a.db.GetStats()
	if err != nil {
		return fail(c, 500, "Failed to fetch stats", err)
	}
	return c.JSON(stats)
}

// Search recherche des appareils par IP
func (a *API) Search(c *fiber.Ctx) error {
	query := c.Query("q")
	if query == "" {
		return fail(c, 400, "Query parameter 'q' is required", nil)
	}
	devices, err := a.db.GetDevices(map[string]interface{}{"ip": query, "limit": 50})
	if err != nil {
		return fail(c, 500, "Search failed", err)
	}
	return c.JSON(fiber.Map{"results": devices, "query": query, "count": len(devices)})
}

// GetAlerts retourne les alertes ouvertes (?all=true pour inclure celles résolues)
func (a *API) GetAlerts(c *fiber.Ctx) error {
	alerts, err := a.db.ListAlerts(c.QueryBool("all"), c.QueryInt("limit", 200))
	if err != nil {
		return fail(c, 500, "Failed to fetch alerts", err)
	}
	return c.JSON(fiber.Map{"alerts": alerts, "count": len(alerts)})
}

// SetAlertRead marque une alerte comme lue / non lue
func (a *API) SetAlertRead(c *fiber.Ctx) error {
	id, err := strconv.Atoi(c.Params("id"))
	if err != nil {
		return fail(c, 400, "Invalid alert id", nil)
	}
	body := struct {
		Read *bool `json:"read"`
	}{}
	if err := c.BodyParser(&body); err != nil || body.Read == nil {
		return fail(c, 400, `Body must be {"read": true|false}`, nil)
	}
	if err := a.db.SetAlertRead(id, *body.Read); err != nil {
		return fail(c, 404, "Alert not found", nil)
	}
	return c.JSON(fiber.Map{"id": id, "read": *body.Read})
}

// MarkAllRead marque toutes les alertes ouvertes comme lues
func (a *API) MarkAllRead(c *fiber.Ctx) error {
	if err := a.db.MarkAllAlertsRead(); err != nil {
		return fail(c, 500, "Failed to update alerts", err)
	}
	return c.JSON(fiber.Map{"ok": true})
}

// RequestScan demande un scan immédiat au scheduler
func (a *API) RequestScan(c *fiber.Ctx) error {
	if err := a.db.RequestScan(); err != nil {
		return fail(c, 500, "Failed to request scan", err)
	}
	return c.Status(202).JSON(fiber.Map{"requested": true})
}

// GetLatestScan retourne le dernier scan et l'état courant
func (a *API) GetLatestScan(c *fiber.Ctx) error {
	scan, err := a.db.LatestScan()
	if err != nil {
		return fail(c, 500, "Failed to fetch scan", err)
	}
	scanning, _ := a.db.ScanInProgress()
	return c.JSON(fiber.Map{"scan": scan, "scanning": scanning})
}
