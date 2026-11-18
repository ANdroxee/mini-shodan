package database

import (
	"github.com/lib/pq"

	"shodan-home/internal/risk"
)

// GetStats calcule les statistiques du réseau
func (db *DB) GetStats() (*Stats, error) {
	stats := &Stats{TopServices: []ServiceCount{}, RiskyPorts: []Port{}}

	if err := db.conn.QueryRow(`SELECT
		(SELECT COUNT(*) FROM devices),
		(SELECT COUNT(*) FROM devices WHERE last_seen > NOW() - INTERVAL '24 hours'),
		(SELECT COUNT(*) FROM ports),
		(SELECT COUNT(*) FROM devices WHERE created_at >= date_trunc('day', NOW()))`).
		Scan(&stats.TotalDevices, &stats.ActiveDevices, &stats.TotalPorts, &stats.NewToday); err != nil {
		return nil, err
	}

	rows, err := db.conn.Query(`
		SELECT service, COUNT(*) FROM ports WHERE COALESCE(service, '') <> ''
		GROUP BY service ORDER BY COUNT(*) DESC LIMIT 10`)
	if err != nil {
		return nil, err
	}
	for rows.Next() {
		var sc ServiceCount
		if err := rows.Scan(&sc.Service, &sc.Count); err == nil {
			stats.TopServices = append(stats.TopServices, sc)
		}
	}
	rows.Close()

	rows, err = db.conn.Query(`
		SELECT id, device_id, port, protocol, COALESCE(service, ''), COALESCE(version, ''), COALESCE(banner, ''), last_seen
		FROM ports WHERE port = ANY($1) ORDER BY last_seen DESC LIMIT 20`, pq.Array(risk.Numbers()))
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	if stats.RiskyPorts, err = collectPorts(rows); err != nil {
		return nil, err
	}
	return stats, nil
}
