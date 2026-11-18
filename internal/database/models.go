package database

import (
	"time"
)

// Device représente un appareil sur le réseau
type Device struct {
	ID        int       `json:"id"`
	IP        string    `json:"ip"`
	Hostname  string    `json:"hostname"`
	MAC       string    `json:"mac"`
	Vendor    string    `json:"vendor"`
	OS        string    `json:"os"`
	Name      string    `json:"name"`  // nom donné par l'utilisateur
	Known     bool      `json:"known"` // l'utilisateur a confirmé connaître l'appareil
	FirstSeen time.Time `json:"first_seen"`
	LastSeen  time.Time `json:"last_seen"`
	Ports     []Port    `json:"ports"`
	// Calculés à la lecture
	Status string `json:"status,omitempty"` // bad, warn, new, ok, off
	Online bool   `json:"online"`
	// Partial : l'hôte a répondu à la découverte mais le scan détaillé a échoué
	// (ses ports ne doivent alors pas être considérés comme fermés).
	Partial bool `json:"-"`
}

// Port représente un port ouvert sur un appareil
type Port struct {
	ID       int       `json:"id"`
	DeviceID int       `json:"device_id"`
	Number   int       `json:"port"`
	Protocol string    `json:"protocol"`
	Service  string    `json:"service"`
	Version  string    `json:"version"`
	Banner   string    `json:"banner,omitempty"`
	LastSeen time.Time `json:"last_seen"`
	Risk     string    `json:"risk,omitempty"` // critical, high, medium
}

// Stats représente les statistiques du réseau
type Stats struct {
	TotalDevices  int            `json:"total_devices"`
	TotalPorts    int            `json:"total_ports"`
	ActiveDevices int            `json:"active_devices"`
	NewToday      int            `json:"new_today"`
	TopServices   []ServiceCount `json:"top_services"`
	RiskyPorts    []Port         `json:"risky_ports"`
}

// ServiceCount compte les occurrences d'un service
type ServiceCount struct {
	Service string `json:"service"`
	Count   int    `json:"count"`
}

// Scan représente une exécution du scanner
type Scan struct {
	ID         int        `json:"id"`
	Target     string     `json:"target"`
	StartTime  time.Time  `json:"start_time"`
	EndTime    *time.Time `json:"end_time"`
	Duration   int        `json:"duration_seconds"`
	HostsFound int        `json:"hosts_found"`
	PortsFound int        `json:"ports_found"`
	Status     string     `json:"status"` // running, completed, failed
}

// Alert représente une alerte
type Alert struct {
	ID         int       `json:"id"`
	Type       string    `json:"type"` // new_device, risky_port, version_change, device_offline
	DeviceID   int       `json:"device_id"`
	DeviceIP   string    `json:"device_ip"`
	DeviceName string    `json:"device_name"`
	Port       *int      `json:"port"`
	Message    string    `json:"message"`
	Details    string    `json:"details"`
	Severity   string    `json:"severity"` // info, medium, high, critical
	CreatedAt  time.Time `json:"created_at"`
	Read       bool      `json:"read"`
	Resolved   bool      `json:"resolved"`
}
