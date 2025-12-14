package database

import (
	"database/sql"
	"fmt"
	"time"

	_ "github.com/lib/pq"
)

type DB struct {
	conn *sql.DB
}

// Connect établit la connexion à PostgreSQL
func Connect(connStr string) (*DB, error) {
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Test de connexion
	if err := conn.Ping(); err != nil {
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	// Configuration du pool de connexions
	conn.SetMaxOpenConns(25)
	conn.SetMaxIdleConns(5)
	conn.SetConnMaxLifetime(5 * time.Minute)

	return &DB{conn: conn}, nil
}

// Close ferme la connexion
func (db *DB) Close() error {
	return db.conn.Close()
}

// SaveDevice enregistre ou met à jour un device
func (db *DB) SaveDevice(device Device) (int, error) {
	var deviceID int

	// Upsert du device
	query := `
		INSERT INTO devices (ip, hostname, mac, vendor, os, last_seen)
		VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (ip) 
		DO UPDATE SET 
			hostname = EXCLUDED.hostname,
			mac = EXCLUDED.mac,
			vendor = EXCLUDED.vendor,
			os = EXCLUDED.os,
			last_seen = EXCLUDED.last_seen
		RETURNING id
	`

	err := db.conn.QueryRow(
		query,
		device.IP,
		device.Hostname,
		device.MAC,
		device.Vendor,
		device.OS,
		device.LastSeen,
	).Scan(&deviceID)

	if err != nil {
		return 0, fmt.Errorf("failed to save device: %w", err)
	}

	// Sauvegarder les ports
	for _, port := range device.Ports {
		if err := db.SavePort(deviceID, port); err != nil {
			return deviceID, fmt.Errorf("failed to save port: %w", err)
		}
	}

	return deviceID, nil
}

// SavePort enregistre ou met à jour un port
func (db *DB) SavePort(deviceID int, port Port) error {
	query := `
		INSERT INTO ports (device_id, port, protocol, service, version, banner, last_seen)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (device_id, port, protocol)
		DO UPDATE SET
			service = EXCLUDED.service,
			version = EXCLUDED.version,
			banner = EXCLUDED.banner,
			last_seen = EXCLUDED.last_seen
	`

	_, err := db.conn.Exec(
		query,
		deviceID,
		port.Number,
		port.Protocol,
		port.Service,
		port.Version,
		port.Banner,
		time.Now(),
	)

	return err
}

// GetDevices récupère tous les devices avec filtres
func (db *DB) GetDevices(filters map[string]interface{}) ([]Device, error) {
	query := `
		SELECT DISTINCT d.id, d.ip, d.hostname, d.mac, d.vendor, d.os, d.last_seen
		FROM devices d
		LEFT JOIN ports p ON d.id = p.device_id
		WHERE 1=1
	`
	args := []interface{}{}
	argCount := 1

	// Filtres dynamiques
	if service, ok := filters["service"].(string); ok && service != "" {
		query += fmt.Sprintf(" AND p.service = $%d", argCount)
		args = append(args, service)
		argCount++
	}

	if ip, ok := filters["ip"].(string); ok && ip != "" {
		query += fmt.Sprintf(" AND d.ip LIKE $%d", argCount)
		args = append(args, "%"+ip+"%")
		argCount++
	}

	if activeOnly, ok := filters["active"].(bool); ok && activeOnly {
		query += fmt.Sprintf(" AND d.last_seen > NOW() - INTERVAL '24 hours'")
	}

	query += " ORDER BY d.last_seen DESC"

	if limit, ok := filters["limit"].(int); ok && limit > 0 {
		query += fmt.Sprintf(" LIMIT $%d", argCount)
		args = append(args, limit)
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var devices []Device
	for rows.Next() {
		var d Device
		err := rows.Scan(&d.ID, &d.IP, &d.Hostname, &d.MAC, &d.Vendor, &d.OS, &d.LastSeen)
		if err != nil {
			continue
		}
		devices = append(devices, d)
	}

	return devices, nil
}

// GetDeviceByIP récupère un device avec ses ports
func (db *DB) GetDeviceByIP(ip string) (*Device, error) {
	var device Device

	query := `
		SELECT id, ip, hostname, mac, vendor, os, last_seen
		FROM devices
		WHERE ip = $1
	`

	err := db.conn.QueryRow(query, ip).Scan(
		&device.ID,
		&device.IP,
		&device.Hostname,
		&device.MAC,
		&device.Vendor,
		&device.OS,
		&device.LastSeen,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("device not found")
	}
	if err != nil {
		return nil, err
	}

	// Récupérer les ports
	ports, err := db.GetPortsForDevice(device.ID)
	if err != nil {
		return nil, err
	}
	device.Ports = ports

	return &device, nil
}

// GetPortsForDevice récupère les ports d'un device
func (db *DB) GetPortsForDevice(deviceID int) ([]Port, error) {
	query := `
		SELECT id, device_id, port, protocol, service, version, banner, last_seen
		FROM ports
		WHERE device_id = $1
		ORDER BY port ASC
	`

	rows, err := db.conn.Query(query, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var ports []Port
	for rows.Next() {
		var p Port
		err := rows.Scan(&p.ID, &p.DeviceID, &p.Number, &p.Protocol, &p.Service, &p.Version, &p.Banner, &p.LastSeen)
		if err != nil {
			continue
		}
		ports = append(ports, p)
	}

	return ports, nil
}

// GetStats calcule les statistiques du réseau
func (db *DB) GetStats() (*Stats, error) {
	stats := &Stats{}

	// Total devices
	db.conn.QueryRow("SELECT COUNT(*) FROM devices").Scan(&stats.TotalDevices)

	// Active devices (vus dans les dernières 24h)
	db.conn.QueryRow(`
		SELECT COUNT(*) FROM devices 
		WHERE last_seen > NOW() - INTERVAL '24 hours'
	`).Scan(&stats.ActiveDevices)

	// Total ports
	db.conn.QueryRow("SELECT COUNT(*) FROM ports").Scan(&stats.TotalPorts)

	// New today
	today := time.Now().Truncate(24 * time.Hour)
	db.conn.QueryRow(`
		SELECT COUNT(*) FROM devices 
		WHERE last_seen >= $1
	`, today).Scan(&stats.NewToday)

	// Top services
	rows, err := db.conn.Query(`
		SELECT service, COUNT(*) as count
		FROM ports
		WHERE service != ''
		GROUP BY service
		ORDER BY count DESC
		LIMIT 10
	`)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var sc ServiceCount
			rows.Scan(&sc.Service, &sc.Count)
			stats.TopServices = append(stats.TopServices, sc)
		}
	}

	// Risky ports
	riskyPorts := []int{21, 23, 445, 3389, 1433, 3306, 5432, 6379}
	query := `
		SELECT id, device_id, port, protocol, service, version, banner, last_seen
		FROM ports
		WHERE port = ANY($1)
		ORDER BY last_seen DESC
		LIMIT 20
	`
	rows, err = db.conn.Query(query, riskyPorts)
	if err == nil {
		defer rows.Close()
		for rows.Next() {
			var p Port
			rows.Scan(&p.ID, &p.DeviceID, &p.Number, &p.Protocol, &p.Service, &p.Version, &p.Banner, &p.LastSeen)
			stats.RiskyPorts = append(stats.RiskyPorts, p)
		}
	}

	return stats, nil
}

// CreateAlert crée une nouvelle alerte
func (db *DB) CreateAlert(alert Alert) error {
	query := `
		INSERT INTO alerts (type, device_id, message, severity, created_at, read)
		VALUES ($1, $2, $3, $4, $5, $6)
	`

	_, err := db.conn.Exec(
		query,
		alert.Type,
		alert.DeviceID,
		alert.Message,
		alert.Severity,
		time.Now(),
		false,
	)

	return err
}

// GetUnreadAlerts récupère les alertes non lues
func (db *DB) GetUnreadAlerts() ([]Alert, error) {
	query := `
		SELECT id, type, device_id, message, severity, created_at, read
		FROM alerts
		WHERE read = false
		ORDER BY created_at DESC
		LIMIT 50
	`

	rows, err := db.conn.Query(query)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var alerts []Alert
	for rows.Next() {
		var a Alert
		err := rows.Scan(&a.ID, &a.Type, &a.DeviceID, &a.Message, &a.Severity, &a.CreatedAt, &a.Read)
		if err != nil {
			continue
		}
		alerts = append(alerts, a)
	}

	return alerts, nil
}
