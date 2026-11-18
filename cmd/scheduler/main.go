// Scheduler : tourne en boucle. Scanne le réseau selon SCAN_SCHEDULE (cron)
// et à chaque fois qu'un scan est demandé depuis l'interface (« Vérifier maintenant »).
package main

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/robfig/cron/v3"

	"shodan-home/internal/database"
	"shodan-home/internal/monitor"
	"shodan-home/internal/scanner"
)

type runner struct {
	db      *database.DB
	scanner *scanner.Scanner
	target  string
	mu      sync.Mutex // un seul scan à la fois
	busy    bool
}

func main() {
	dbURL := getEnv("DATABASE_URL", "postgres://shodan:shodan123@localhost:5432/shodan_home?sslmode=disable")
	target := getEnv("SCAN_TARGET", "192.168.1.0/24")
	schedule := getEnv("SCAN_SCHEDULE", "0 3 * * *") // 3h du matin par défaut
	workers := getEnvInt("WORKERS", 10)
	loc, err := time.LoadLocation(getEnv("SCAN_TIMEZONE", "Europe/Paris"))
	if err != nil {
		log.Fatalf("❌ Invalid SCAN_TIMEZONE: %v", err)
	}
	monitor.Location = loc

	fmt.Println("⏰ Shodan@Home Scheduler")
	fmt.Printf("Target: %s · Schedule: %s (%s) · Workers: %d\n\n", target, schedule, loc, workers)

	db, err := database.Connect(dbURL)
	if err != nil {
		log.Fatalf("❌ Database connection failed: %v", err)
	}
	defer db.Close()
	if err := db.Migrate(); err != nil {
		log.Fatalf("❌ %v", err)
	}
	db.MarkInterruptedScans()

	r := &runner{db: db, scanner: scanner.NewScanner(workers), target: target}

	c := cron.New(cron.WithLocation(loc))
	if _, err := c.AddFunc(schedule, func() { r.start("planifié") }); err != nil {
		log.Fatalf("❌ Invalid SCAN_SCHEDULE %q: %v", schedule, err)
	}
	c.Start()
	fmt.Printf("✅ Scheduler started (%s)\n", schedule)

	if getEnv("RUN_ON_START", "false") == "true" {
		r.start("démarrage")
	}

	// Scans demandés depuis l'interface
	stop := make(chan struct{})
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for {
			select {
			case <-stop:
				return
			case <-ticker.C:
				if r.isBusy() {
					continue
				}
				if ok, err := db.TakeScanRequest(); err == nil && ok {
					r.start("demandé")
				}
			}
		}
	}()

	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	<-sig
	fmt.Println("\n🛑 Stopping scheduler...")
	close(stop)
	<-c.Stop().Done()
}

func (r *runner) isBusy() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.busy
}

// start lance un scan en arrière-plan, sauf si un scan tourne déjà
func (r *runner) start(reason string) {
	r.mu.Lock()
	if r.busy {
		r.mu.Unlock()
		fmt.Printf("⏭️  Scan %s ignoré : un scan est déjà en cours\n", reason)
		return
	}
	r.busy = true
	r.mu.Unlock()

	go func() {
		defer func() {
			r.mu.Lock()
			r.busy = false
			r.mu.Unlock()
		}()
		line := strings.Repeat("=", 50)
		fmt.Printf("\n%s\n🔍 Scan %s — %s\n%s\n", line, reason, time.Now().Format("2006-01-02 15:04:05"), line)

		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Minute)
		defer cancel()
		res, err := monitor.Run(ctx, r.db, r.scanner, r.target)
		if err != nil {
			log.Printf("❌ Scan failed: %v", err)
			return
		}
		fmt.Printf("✅ %d devices, %d ports, %d new alerts in %v\n\n",
			res.Devices, res.Ports, res.Alerts, res.Duration.Round(time.Second))
	}()
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
