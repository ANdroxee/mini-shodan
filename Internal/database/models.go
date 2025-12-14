package database

import (
	"time"
)

// Device représente un appareil sur le réseau
type Device struct {
	ID       int       `json:"id"`
	IP       string    `json:"ip"`
	Hostname string    `json:"hostname"`
	MAC      string    `json:"mac"`
	Vendor   string    `json:"vendor"`
	OS       string    `json:"os"`
	LastSeen time.Time `json:"last_seen"`
	Ports    []Port    `json:"ports,omitempty"`
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

// ScanResult représente le résultat d'un scan
type ScanResult struct {
	ID         int       `json:"id"`
	Target     string    `json:"target"`
	StartTime  time.Time `json:"start_time"`
	EndTime    time.Time `json:"end_time"`
	Duration   int       `json:"duration_seconds"`
	HostsFound int       `json:"hosts_found"`
	PortsFound int       `json:"ports_found"`
}

// Alert représente une alerte système
type Alert struct {
	ID        int       `json:"id"`
	Type      string    `json:"type"` // new_device, risky_port, service_change
	DeviceID  int       `json:"device_id"`
	Message   string    `json:"message"`
	Severity  string    `json:"severity"` // low, medium, high, critical
	CreatedAt time.Time `json:"created_at"`
	Read      bool      `json:"read"`
}
