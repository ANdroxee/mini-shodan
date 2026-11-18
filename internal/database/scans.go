package database

import (
	"database/sql"
	"time"
)

const scanColumns = `id, target, start_time, end_time, COALESCE(duration_seconds, 0), COALESCE(hosts_found, 0),
	COALESCE(ports_found, 0), COALESCE(status, 'running')`

func readScan(row *sql.Row) (*Scan, error) {
	var s Scan
	var end sql.NullTime
	err := row.Scan(&s.ID, &s.Target, &s.StartTime, &end, &s.Duration, &s.HostsFound, &s.PortsFound, &s.Status)
	if err == sql.ErrNoRows {
		return nil, nil
	}
	if err != nil {
		return nil, err
	}
	if end.Valid {
		s.EndTime = &end.Time
	}
	return &s, nil
}

// StartScan enregistre le début d'un scan
func (db *DB) StartScan(target string, start time.Time) (int, error) {
	var id int
	err := db.conn.QueryRow(`INSERT INTO scans (target, start_time, status) VALUES ($1, $2, 'running') RETURNING id`,
		target, start).Scan(&id)
	return id, err
}

// FinishScan enregistre la fin d'un scan
func (db *DB) FinishScan(id, hosts, ports int, status string) error {
	_, err := db.conn.Exec(`
		UPDATE scans SET end_time = NOW(), duration_seconds = EXTRACT(EPOCH FROM (NOW() - start_time))::int,
			hosts_found = $2, ports_found = $3, status = $4
		WHERE id = $1`, id, hosts, ports, status)
	return err
}

// LastCompletedScan renvoie le dernier scan terminé avec succès (hors scan `excludeID`)
func (db *DB) LastCompletedScan(excludeID int) (*Scan, error) {
	return readScan(db.conn.QueryRow(`SELECT `+scanColumns+` FROM scans
		WHERE status = 'completed' AND id <> $1 ORDER BY start_time DESC LIMIT 1`, excludeID))
}

// LatestScan renvoie le scan le plus récent, quel que soit son état
func (db *DB) LatestScan() (*Scan, error) {
	return readScan(db.conn.QueryRow(`SELECT ` + scanColumns + ` FROM scans ORDER BY start_time DESC LIMIT 1`))
}

// MarkInterruptedScans clôt les scans restés « running » (scheduler arrêté en plein scan)
func (db *DB) MarkInterruptedScans() error {
	_, err := db.conn.Exec(`UPDATE scans SET status = 'failed', end_time = NOW() WHERE status = 'running'`)
	return err
}

// RequestScan demande un scan immédiat (pris en charge par le scheduler)
func (db *DB) RequestScan() error {
	_, err := db.conn.Exec(`
		INSERT INTO scan_requests (requested_at)
		SELECT NOW() WHERE NOT EXISTS (SELECT 1 FROM scan_requests WHERE picked_at IS NULL)`)
	return err
}

// TakeScanRequest prend la plus ancienne demande en attente ; renvoie false s'il n'y en a pas
func (db *DB) TakeScanRequest() (bool, error) {
	var id int
	err := db.conn.QueryRow(`
		UPDATE scan_requests SET picked_at = NOW()
		WHERE id = (SELECT id FROM scan_requests WHERE picked_at IS NULL ORDER BY id LIMIT 1 FOR UPDATE SKIP LOCKED)
		RETURNING id`).Scan(&id)
	if err == sql.ErrNoRows {
		return false, nil
	}
	return err == nil, err
}

// ScanInProgress indique si un scan tourne ou est demandé
func (db *DB) ScanInProgress() (bool, error) {
	var n int
	err := db.conn.QueryRow(`
		SELECT (SELECT COUNT(*) FROM scan_requests WHERE picked_at IS NULL)
		     + (SELECT COUNT(*) FROM scans WHERE status = 'running' AND start_time > NOW() - INTERVAL '2 hours')`).Scan(&n)
	return n > 0, err
}
