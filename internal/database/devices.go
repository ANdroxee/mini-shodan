package database

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"shodan-home/internal/risk"
)

const deviceColumns = `id, ip, COALESCE(hostname, ''), COALESCE(mac, ''), COALESCE(vendor, ''), COALESCE(os, ''),
	COALESCE(name, ''), known, created_at, last_seen`

func scanDevice(row interface{ Scan(...any) error }, d *Device) error {
	return row.Scan(&d.ID, &d.IP, &d.Hostname, &d.MAC, &d.Vendor, &d.OS, &d.Name, &d.Known, &d.FirstSeen, &d.LastSeen)
}

// FindDevice renvoie l'appareil d'une IP, ou nil s'il n'existe pas encore
func (db *DB) FindDevice(ip string) (*Device, error) {
	var d Device
	err := scanDevice(db.conn.QueryRow(`SELECT `+deviceColumns+` FROM devices WHERE ip = $1`, ip), &d)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	return &d, nil
}

// UpsertDevice enregistre un appareil vu lors d'un scan, sans écraser des infos connues par des valeurs vides
// ni le nom / statut donnés par l'utilisateur.
func (db *DB) UpsertDevice(d Device, seenAt time.Time) (int, error) {
	var id int
	err := db.conn.QueryRow(`
		INSERT INTO devices (ip, hostname, mac, vendor, os, last_seen)
		VALUES ($1, NULLIF($2, ''), NULLIF($3, ''), NULLIF($4, ''), NULLIF($5, ''), $6)
		ON CONFLICT (ip) DO UPDATE SET
			hostname  = COALESCE(EXCLUDED.hostname, devices.hostname),
			mac       = COALESCE(EXCLUDED.mac, devices.mac),
			vendor    = COALESCE(EXCLUDED.vendor, devices.vendor),
			os        = COALESCE(EXCLUDED.os, devices.os),
			last_seen = EXCLUDED.last_seen
		RETURNING id`,
		d.IP, d.Hostname, d.MAC, d.Vendor, d.OS, seenAt,
	).Scan(&id)
	if err != nil {
		return 0, fmt.Errorf("failed to save device %s: %w", d.IP, err)
	}
	return id, nil
}

// SaveDevice : compatibilité avec l'ancien code (enregistre l'appareil et ses ports)
func (db *DB) SaveDevice(d Device) (int, error) {
	now := time.Now().UTC()
	id, err := db.UpsertDevice(d, now)
	if err != nil {
		return 0, err
	}
	for _, p := range d.Ports {
		if err := db.UpsertPort(id, p, now); err != nil {
			return id, err
		}
	}
	return id, nil
}

// UpsertPort enregistre ou met à jour un port ouvert
func (db *DB) UpsertPort(deviceID int, p Port, seenAt time.Time) error {
	if p.Protocol == "" {
		p.Protocol = "tcp"
	}
	_, err := db.conn.Exec(`
		INSERT INTO ports (device_id, port, protocol, service, version, banner, last_seen)
		VALUES ($1, $2, $3, $4, $5, $6, $7)
		ON CONFLICT (device_id, port, protocol) DO UPDATE SET
			service = EXCLUDED.service,
			version = EXCLUDED.version,
			banner = EXCLUDED.banner,
			last_seen = EXCLUDED.last_seen`,
		deviceID, p.Number, p.Protocol, p.Service, strings.TrimSpace(p.Version), p.Banner, seenAt,
	)
	return err
}

// DeleteStalePorts supprime les ports d'un appareil qui n'ont pas été revus depuis `before`
// (ils ont été fermés) et les renvoie.
func (db *DB) DeleteStalePorts(deviceID int, before time.Time) ([]Port, error) {
	rows, err := db.conn.Query(`
		DELETE FROM ports WHERE device_id = $1 AND last_seen < $2
		RETURNING id, device_id, port, protocol, COALESCE(service, ''), COALESCE(version, ''), COALESCE(banner, ''), last_seen`,
		deviceID, before)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectPorts(rows)
}

func collectPorts(rows *sql.Rows) ([]Port, error) {
	ports := []Port{}
	for rows.Next() {
		var p Port
		if err := rows.Scan(&p.ID, &p.DeviceID, &p.Number, &p.Protocol, &p.Service, &p.Version, &p.Banner, &p.LastSeen); err != nil {
			return nil, err
		}
		if _, sev, ok := risk.Port(p.Number); ok {
			p.Risk = sev
		}
		ports = append(ports, p)
	}
	return ports, rows.Err()
}

// GetPortsForDevice récupère les ports ouverts d'un appareil
func (db *DB) GetPortsForDevice(deviceID int) ([]Port, error) {
	rows, err := db.conn.Query(`
		SELECT id, device_id, port, protocol, COALESCE(service, ''), COALESCE(version, ''), COALESCE(banner, ''), last_seen
		FROM ports WHERE device_id = $1 ORDER BY port ASC`, deviceID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	return collectPorts(rows)
}

// GetDevices récupère les appareils (avec leurs ports) selon des filtres optionnels
func (db *DB) GetDevices(filters map[string]interface{}) ([]Device, error) {
	query := `SELECT ` + deviceColumns + ` FROM devices d WHERE 1=1`
	args := []interface{}{}

	if service, ok := filters["service"].(string); ok && service != "" {
		args = append(args, service)
		query += fmt.Sprintf(" AND EXISTS (SELECT 1 FROM ports p WHERE p.device_id = d.id AND p.service = $%d)", len(args))
	}
	if ip, ok := filters["ip"].(string); ok && ip != "" {
		args = append(args, "%"+ip+"%")
		query += fmt.Sprintf(" AND d.ip LIKE $%d", len(args))
	}
	if activeOnly, ok := filters["active"].(bool); ok && activeOnly {
		query += " AND d.last_seen > NOW() - INTERVAL '24 hours'"
	}
	query += " ORDER BY d.last_seen DESC"
	if limit, ok := filters["limit"].(int); ok && limit > 0 {
		args = append(args, limit)
		query += fmt.Sprintf(" LIMIT $%d", len(args))
	}

	rows, err := db.conn.Query(query, args...)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	devices := []Device{}
	for rows.Next() {
		var d Device
		if err := scanDevice(rows, &d); err != nil {
			return nil, err
		}
		devices = append(devices, d)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	for i := range devices {
		if devices[i].Ports, err = db.GetPortsForDevice(devices[i].ID); err != nil {
			return nil, err
		}
	}
	return devices, nil
}

// GetDeviceByIP récupère un appareil avec ses ports
func (db *DB) GetDeviceByIP(ip string) (*Device, error) {
	d, err := db.FindDevice(ip)
	if err != nil {
		return nil, err
	}
	if d == nil {
		return nil, fmt.Errorf("device not found")
	}
	if d.Ports, err = db.GetPortsForDevice(d.ID); err != nil {
		return nil, err
	}
	return d, nil
}

// SetDeviceName donne un nom à un appareil (vide = revenir au nom automatique)
func (db *DB) SetDeviceName(ip, name string) error {
	res, err := db.conn.Exec(`UPDATE devices SET name = NULLIF($2, '') WHERE ip = $1`, ip, name)
	return checkAffected(res, err)
}

// SetDeviceKnown marque un appareil comme connu ; l'alerte « nouvel appareil » est alors close
func (db *DB) SetDeviceKnown(ip string, known bool) error {
	res, err := db.conn.Exec(`UPDATE devices SET known = $2 WHERE ip = $1`, ip, known)
	if err := checkAffected(res, err); err != nil {
		return err
	}
	if known {
		_, err = db.conn.Exec(`
			UPDATE alerts SET resolved = TRUE, read = TRUE
			WHERE type = 'new_device' AND device_id = (SELECT id FROM devices WHERE ip = $1)`, ip)
	}
	return err
}

func checkAffected(res sql.Result, err error) error {
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("device not found")
	}
	return nil
}

// OfflineSince renvoie les appareils vus pendant le scan précédent mais pas pendant le scan courant
func (db *DB) OfflineSince(prevStart, currentStart time.Time) ([]Device, error) {
	rows, err := db.conn.Query(`SELECT `+deviceColumns+` FROM devices WHERE last_seen >= $1 AND last_seen < $2`, prevStart, currentStart)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	out := []Device{}
	for rows.Next() {
		var d Device
		if err := scanDevice(rows, &d); err != nil {
			return nil, err
		}
		out = append(out, d)
	}
	return out, rows.Err()
}
