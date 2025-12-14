package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"shodan-home/internal/database"
	"shodan-home/internal/scanner"

	"github.com/robfig/cron/v3"
)

func main() {
	// Configuration
	dbURL := getEnv("DATABASE_URL", "postgres://shodan:shodan123@localhost:5432/shodan_home?sslmode=disable")
	target := getEnv("SCAN_TARGET", "192.168.1.0/24")
	schedule := getEnv("SCAN_SCHEDULE", "0 3 * * *") // 3h du matin par défaut
	workers := getEnvInt("WORKERS", 10)

	fmt.Println("⏰ Shodan@Home Scheduler")
	fmt.Println("========================")
	fmt.Printf("Target: %s\n", target)
	fmt.Printf("Schedule: %s\n", schedule)
	fmt.Printf("Workers: %d\n\n", workers)

	// Connexion DB
	db, err := database.Connect(dbURL)
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer db.Close()
	fmt.Println("✅ Database connected")

	// Créer le scanner
	s := scanner.NewScanner(workers)

	// Configurer le cron
	c := cron.New()

	_, err = c.AddFunc(schedule, func() {
		runScheduledScan(db, s, target)
	})

	if err != nil {
		log.Fatalf("❌ Failed to schedule job: %v", err)
	}

	// Démarrer le cron
	c.Start()
	fmt.Printf("✅ Scheduler started with cron: %s\n", schedule)
	fmt.Println("⏳ Waiting for scheduled scans...")
	fmt.Println("💡 Press Ctrl+C to stop")

	// Option: lancer un scan immédiat au démarrage
	if getEnv("RUN_ON_START", "false") == "true" {
		fmt.Println("\n🚀 Running initial scan...")
		go runScheduledScan(db, s, target)
	}

	// Graceful shutdown
	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
	<-sigChan

	fmt.Println("\n🛑 Stopping scheduler...")
	c.Stop()
	fmt.Println("✅ Scheduler stopped")
}

func runScheduledScan(db *database.DB, s *scanner.Scanner, target string) {
	fmt.Printf("\n" + "=".repeat(50) + "\n")
	fmt.Printf("🔍 Starting scheduled scan at %s\n", time.Now().Format("2006-01-02 15:04:05"))
	fmt.Printf("=".repeat(50) + "\n")

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
	defer cancel()

	start := time.Now()
	devices, err := s.ScanNetwork(ctx, target)
	duration := time.Since(start)

	if err != nil {
		log.Printf("❌ Scan failed: %v", err)
		return
	}

	fmt.Printf("✅ Scan completed in %v\n", duration)
	fmt.Printf("📊 Found %d devices\n", len(devices))

	// Sauvegarder
	saved := 0
	for _, device := range devices {
		_, err := db.SaveDevice(device)
		if err != nil {
			continue
		}
		saved++
	}

	fmt.Printf("💾 Saved %d/%d devices\n", saved, len(devices))
	fmt.Printf("⏰ Next scan will run according to schedule\n")
	fmt.Printf("=" + repeat(50) + "\n\n")
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

func repeat(n int) string {
	result := ""
	for i := 0; i < n; i++ {
		result += "="
	}
	return result
}
