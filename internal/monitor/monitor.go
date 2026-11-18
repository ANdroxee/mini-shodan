// Package monitor exécute un scan complet du réseau, l'enregistre et produit les alertes.
// Il est utilisé à la fois par le scan ponctuel (cmd/scanner) et par la boucle du scheduler.
package monitor

import (
	"context"
	"fmt"
	"strings"
	"time"

	"shodan-home/internal/database"
	"shodan-home/internal/risk"
	"shodan-home/internal/scanner"
)

// Location : fuseau utilisé pour les dates écrites dans les textes d'alerte
var Location = time.UTC

// Result résume un scan
type Result struct {
	Devices  int
	Ports    int
	Alerts   int
	Duration time.Duration
}

// Run scanne `target`, met la base à jour et crée les alertes :
//   - new_device      : un appareil jamais vu apparaît (sauf au tout premier scan, qui sert de référence)
//   - risky_port      : un port risqué vient de s'ouvrir ; l'alerte se clôt toute seule quand il se ferme
//   - version_change  : le logiciel derrière un port a changé de version
//   - device_offline  : un appareil vu au scan précédent n'a pas répondu
func Run(ctx context.Context, db *database.DB, s *scanner.Scanner, target string) (*Result, error) {
	start := time.Now().UTC()
	scanID, err := db.StartScan(target, start)
	if err != nil {
		return nil, fmt.Errorf("cannot record scan: %w", err)
	}

	devices, err := s.ScanNetwork(ctx, target)
	if err != nil {
		db.FinishScan(scanID, 0, 0, "failed")
		return nil, err
	}

	previous, err := db.LastCompletedScan(scanID)
	if err != nil {
		db.FinishScan(scanID, 0, 0, "failed")
		return nil, err
	}
	baseline := previous == nil // premier scan : on apprend le réseau sans crier « nouveau » partout

	res := &Result{Devices: len(devices)}
	for _, d := range devices {
		n, err := processDevice(db, d, start, baseline)
		if err != nil {
			fmt.Printf("  ⚠️  %s: %v\n", d.IP, err)
			continue
		}
		res.Alerts += n
		res.Ports += len(d.Ports)
	}

	if previous != nil {
		gone, err := db.OfflineSince(previous.StartTime, start)
		if err == nil {
			for _, d := range gone {
				if open, _ := db.HasOpenAlert(d.ID, "device_offline", nil); open {
					continue
				}
				db.CreateAlert(database.Alert{
					Type: "device_offline", DeviceID: d.ID, Severity: "info",
					Message: fmt.Sprintf("Device %s did not answer", d.IP),
					Details: fmt.Sprintf("IP %s · dernière réponse %s\ntype: device_offline · severity: info", d.IP, d.LastSeen.In(Location).Format("02/01 15:04")),
				})
				res.Alerts++
			}
		}
	}

	res.Duration = time.Since(start)
	if err := db.FinishScan(scanID, res.Devices, res.Ports, "completed"); err != nil {
		return res, err
	}
	return res, nil
}

func processDevice(db *database.DB, d database.Device, scanStart time.Time, baseline bool) (int, error) {
	alerts := 0
	existing, err := db.FindDevice(d.IP)
	if err != nil {
		return 0, err
	}

	var before []database.Port
	if existing != nil {
		if before, err = db.GetPortsForDevice(existing.ID); err != nil {
			return 0, err
		}
	}
	oldPorts := map[int]database.Port{}
	for _, p := range before {
		oldPorts[p.Number] = p
	}

	id, err := db.UpsertDevice(d, scanStart)
	if err != nil {
		return 0, err
	}

	// L'appareil répond de nouveau : on clôt son alerte « hors ligne »
	db.ResolveAlerts(id, "device_offline", nil)

	if existing == nil && !baseline {
		db.CreateAlert(database.Alert{
			Type: "new_device", DeviceID: id, Severity: "info",
			Message: fmt.Sprintf("New device %s (%s)", d.IP, orDash(d.Vendor)),
			Details: evidence(d, d.Ports, "new_device", "info"),
		})
		alerts++
	}

	if d.Partial {
		return alerts, nil // scan détaillé raté : on ne conclut rien sur les ports
	}

	for _, p := range d.Ports {
		if err := db.UpsertPort(id, p, scanStart); err != nil {
			return alerts, err
		}
		old, had := oldPorts[p.Number]
		service, severity, risky := risk.Port(p.Number)
		port := p.Number

		if risky {
			if open, _ := db.HasOpenAlert(id, "risky_port", &port); !open {
				db.CreateAlert(database.Alert{
					Type: "risky_port", DeviceID: id, Port: &port, Severity: severity,
					Message: fmt.Sprintf("Risky port %d (%s) open on %s", p.Number, service, d.IP),
					Details: evidence(d, []database.Port{p}, "risky_port", severity),
				})
				alerts++
			}
		}

		newV, oldV := strings.TrimSpace(p.Version), strings.TrimSpace(old.Version)
		if had && newV != "" && oldV != "" && newV != oldV {
			db.CreateAlert(database.Alert{
				Type: "version_change", DeviceID: id, Port: &port, Severity: "info",
				Message: fmt.Sprintf("Port %d on %s: %s → %s", p.Number, d.IP, oldV, newV),
				Details: fmt.Sprintf("IP %s\n%d/%s  avant : %s · après : %s\ntype: version_change · severity: info",
					d.IP, p.Number, p.Protocol, oldV, newV),
			})
			alerts++
		}
	}

	// Ports fermés depuis le scan précédent : leurs alertes se résolvent toutes seules
	closed, err := db.DeleteStalePorts(id, scanStart)
	if err != nil {
		return alerts, err
	}
	for _, p := range closed {
		port := p.Number
		db.ResolveAlerts(id, "risky_port", &port)
	}
	return alerts, nil
}

// evidence produit le bloc « détails techniques » d'une alerte, façon sortie nmap
func evidence(d database.Device, ports []database.Port, alertType, severity string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "IP %s", d.IP)
	if d.MAC != "" {
		fmt.Fprintf(&b, " · MAC %s", d.MAC)
		if d.Vendor != "" {
			fmt.Fprintf(&b, " (%s)", d.Vendor)
		}
	}
	if d.OS != "" {
		fmt.Fprintf(&b, " · %s", d.OS)
	}
	b.WriteString("\n")
	for _, p := range ports {
		fmt.Fprintf(&b, "%-9s open  %-14s %s\n", fmt.Sprintf("%d/%s", p.Number, p.Protocol), p.Service, p.Version)
	}
	fmt.Fprintf(&b, "type: %s · severity: %s", alertType, severity)
	return b.String()
}

func orDash(s string) string {
	if s == "" {
		return "fabricant inconnu"
	}
	return s
}
