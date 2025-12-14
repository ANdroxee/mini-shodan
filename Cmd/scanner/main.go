package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"shodan-home/internal/database"
	"shodan-home/internal/scanner"
)

func main() {
	// Configuration
	dbURL := getEnv("DATABASE_URL", "postgres://shodan:shodan123@localhost:5432/shodan_home?sslmode=disable")
	target := getEnv("SCAN_TARGET", "192.168.1.0/24")
	workers := getEnvInt("WORKERS", 10)

	fmt.Println("🎯 Shodan@Home Scanner")
	fmt.Println("======================")
	fmt.Printf("Target: %s\n", target)
	fmt.Printf("Workers: %d\n\n", workers)

	// Connexion DB
	fmt.Println("🔌 Connecting to database...")
	db, err := database.Connect(dbURL)
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer db.Close()
	fmt.Println("✅ Database connected\n")

	// Créer le scanner
	s := scanner.NewScanner(workers)

	// Lancer le scan
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	fmt.Printf("🔍 Starting network scan...\n")
	start := time.Now()

	devices, err := s.ScanNetwork(ctx, target)
	if err != nil {
		log.Fatalf("❌ Scan failed: %v", err)
	}

	duration := time.Since(start)

	fmt.Printf("\n✅ Scan completed in %v\n", duration)
	fmt.Printf("📊 Found %d devices\n\n", len(devices))

	// Sauvegarder en base
	fmt.Println("💾 Saving results to database...")
	saved := 0
	totalPorts := 0

	for _, device := range devices {
		deviceID, err := db.SaveDevice(device)
		if err != nil {
			fmt.Printf("⚠️  Failed to save %s: %v\n", device.IP, err)
			continue
		}

		saved++
		totalPorts += len(device.Ports)

		// Créer des alertes pour nouveaux devices ou ports risqués
		checkForAlerts(db, deviceID, device)

		fmt.Printf("  ✓ %s (%s) - %d ports\n", device.IP, device.Hostname, len(device.Ports))
	}

	fmt.Printf("\n🎉 Successfully saved %d/%d devices\n", saved, len(devices))
	fmt.Printf("📡 Total ports found: %d\n", totalPorts)
}

func checkForAlerts(db *database.DB, deviceID int, device database.Device) {
	// Ports risqués
	riskyPorts := map[int]string{
		21:   "FTP",
		23:   "Telnet",
		445:  "SMB",
		3389: "RDP",
		1433: "MSSQL",
		3306: "MySQL",
		5432: "PostgreSQL",
		6379: "Redis",
	}

	for _, port := range device.Ports {
		if serviceName, isRisky := riskyPorts[port.Number]; isRisky {
			alert := database.Alert{
				Type:     "risky_port",
				DeviceID: deviceID,
				Message:  fmt.Sprintf("Risky port %d (%s) open on %s", port.Number, serviceName, device.IP),
				Severity: "high",
			}
			db.CreateAlert(alert)
		}
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

func getEnvInt(key string, defaultValue int) int {
	if value := os.Getenv(key); value != "" {
		var result int
		fmt.Sscanf(value, "%d", &result)
		return result
	}
	return defaultValue
}
