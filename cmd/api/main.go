package main

import (
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"shodan-home/internal/api"
	"shodan-home/internal/database"
)

func main() {
	dbURL := getEnv("DATABASE_URL", "postgres://shodan:shodan123@localhost:5432/shodan_home?sslmode=disable")
	port := getEnv("PORT", "8080")

	fmt.Println("🔌 Connecting to database...")
	db, err := database.Connect(dbURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()
	if err := db.Migrate(); err != nil {
		log.Fatalf("❌ %v", err)
	}
	fmt.Println("✅ Database ready")

	app := api.NewAPI(db, api.Config{
		Target:   getEnv("SCAN_TARGET", "192.168.1.0/24"),
		Schedule: getEnv("SCAN_SCHEDULE", "0 3 * * *"),
		Timezone: getEnv("SCAN_TIMEZONE", "Europe/Paris"),
	}).SetupRoutes()

	go func() {
		sig := make(chan os.Signal, 1)
		signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
		<-sig
		fmt.Println("\n🛑 Shutting down gracefully...")
		app.Shutdown()
	}()

	fmt.Printf("🚀 API server on :%s (health: /health)\n", port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

func getEnv(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
