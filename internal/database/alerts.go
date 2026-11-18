package database

import (
	"database/sql"
	"fmt"
	"time"
)

// CreateAlert crée une nouvelle alerte
func (db *DB) CreateAlert(a Alert) error {
	_, err := db.conn.Exec(`
		INSERT INTO alerts (type, device_id, port, message, details, severity, created_at, read, resolved)
		VALUES ($1, $2, $3, $4, $5, $6, $7, FALSE, FALSE)`,
		a.Type, a.DeviceID, a.Port, a.Message, a.Details, a.Severity, time.Now().UTC(),
	)
	return err
}

// HasOpenAlert indique si une alerte non résolue existe déjà (évite les doublons à chaque scan)
func (db *DB) HasOpenAlert(deviceID int, alertType string, port *int) (bool, error) {
	var n int
	err := db.conn.QueryRow(`
		SELECT COUNT(*) FROM alerts
		WHERE device_id = $1 AND type = $2 AND resolved = FALSE AND port IS NOT DISTINCT FROM $3`,
		deviceID, alertType, port).Scan(&n)
	return n > 0, err
}

// ResolveAlerts clôt les alertes d'un type pour un appareil (et un port si précisé)
func (db *DB) ResolveAlerts(deviceID int, alertType string, port *int) error {
	_, err := db.conn.Exec(`
		UPDATE alerts SET resolved = TRUE
		WHERE device_id = $1 AND type = $2 AND resolved = FALSE AND ($3::int IS NULL OR port = $3)`,
		deviceID, alertType, port)
	return err
}

// ListAlerts renvoie les alertes, les plus récentes d'abord. Par défaut uniquement celles non résolues.
func (db *DB) ListAlerts(includeResolved bool, limit int) ([]Alert, error) {
	query := `
		SELECT a.id, a.type, COALESCE(a.device_id, 0), COALESCE(d.ip, ''), COALESCE(d.name, d.hostname, ''),
		       a.port, a.message, COALESCE(a.details, ''), a.severity, a.created_at, COALESCE(a.read, FALSE), a.resolved
		FROM alerts a LEFT JOIN devices d ON d.id = a.device_id`
	if !includeResolved {
		query += ` WHERE a.resolved = FALSE`
	}
	query += ` ORDER BY a.created_at DESC, a.id DESC LIMIT $1`

	rows, err := db.conn.Query(query, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	alerts := []Alert{}
	for rows.Next() {
		var a Alert
		var port sql.NullInt64
		if err := rows.Scan(&a.ID, &a.Type, &a.DeviceID, &a.DeviceIP, &a.DeviceName, &port, &a.Message, &a.Details,
			&a.Severity, &a.CreatedAt, &a.Read, &a.Resolved); err != nil {
			return nil, err
		}
		if port.Valid {
			p := int(port.Int64)
			a.Port = &p
		}
		alerts = append(alerts, a)
	}
	return alerts, rows.Err()
}

// GetUnreadAlerts : compatibilité, alertes non lues et non résolues
func (db *DB) GetUnreadAlerts() ([]Alert, error) {
	all, err := db.ListAlerts(false, 50)
	if err != nil {
		return nil, err
	}
	out := []Alert{}
	for _, a := range all {
		if !a.Read {
			out = append(out, a)
		}
	}
	return out, nil
}

// SetAlertRead marque une alerte comme lue ou non lue
func (db *DB) SetAlertRead(id int, read bool) error {
	res, err := db.conn.Exec(`UPDATE alerts SET read = $2 WHERE id = $1`, id, read)
	if err != nil {
		return err
	}
	if n, _ := res.RowsAffected(); n == 0 {
		return fmt.Errorf("alert not found")
	}
	return nil
}

// MarkAllAlertsRead marque toutes les alertes ouvertes comme lues
func (db *DB) MarkAllAlertsRead() error {
	_, err := db.conn.Exec(`UPDATE alerts SET read = TRUE WHERE resolved = FALSE`)
	return err
}

// OpenAlertsByDevice regroupe les types d'alertes ouvertes par appareil (pour calculer les statuts)
func (db *DB) OpenAlertsByDevice() (map[int][]Alert, error) {
	alerts, err := db.ListAlerts(false, 1000)
	if err != nil {
		return nil, err
	}
	out := map[int][]Alert{}
	for _, a := range alerts {
		out[a.DeviceID] = append(out[a.DeviceID], a)
	}
	return out, nil
}
