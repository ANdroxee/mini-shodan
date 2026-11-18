// Scan ponctuel : un passage sur tout le réseau, puis on s'arrête.
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"time"

	"shodan-home/internal/database"
	"shodan-home/internal/monitor"
	"shodan-home/internal/scanner"
)

func main() {
	dbURL := getEnv("DATABASE_URL", "postgres://shodan:shodan123@localhost:5432/shodan_home?sslmode=disable")
	target := getEnv("SCAN_TARGET", "192.168.1.0/24")
	workers := getEnvInt("WORKERS", 10)
	if loc, err := time.LoadLocation(getEnv("SCAN_TIMEZONE", "Europe/Paris")); err == nil {
		monitor.Location = loc
	}

	fmt.Println("🎯 Shodan@Home Scanner")
	fmt.Printf("Target: %s · Workers: %d\n\n", target, workers)

	db, err := database.Connect(dbURL)
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer db.Close()
	if err := db.Migrate(); err != nil {
		log.Fatalf("❌ %v", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	res, err := monitor.Run(ctx, db, scanner.NewScanner(workers), target)
	if err != nil {
		log.Fatalf("❌ Scan failed: %v", err)
	}
	fmt.Printf("\n✅ Scan completed in %v — %d devices, %d ports, %d new alerts\n",
		res.Duration.Round(time.Second), res.Devices, res.Ports, res.Alerts)
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func getEnvInt(key string, def int) int {
	if v, err := strconv.Atoi(os.Getenv(key)); err == nil && v > 0 {
		return v
	}
	return def
}
