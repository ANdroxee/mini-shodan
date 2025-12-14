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
	// Récupérer la config depuis les variables d'environnement
	dbURL := getEnv("DATABASE_URL", "postgres://shodan:shodan123@localhost:5432/shodan_home?sslmode=disable")
	port := getEnv("PORT", "8080")

	// Connexion à la base de données
	fmt.Println("🔌 Connecting to database...")
	db, err := database.Connect(dbURL)
	if err != nil {
		log.Fatalf("❌ Failed to connect to database: %v", err)
	}
	defer db.Close()
	fmt.Println("✅ Database connected")

	// Créer l'API
	apiHandler := api.NewAPI(db)
	app := apiHandler.SetupRoutes()

	// Graceful shutdown
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)
		<-sigChan

		fmt.Println("\n🛑 Shutting down gracefully...")
		app.Shutdown()
	}()

	// Démarrer le serveur
	fmt.Printf("🚀 API server starting on :%s\n", port)
	fmt.Printf("📡 Health check: http://localhost:%s/health\n", port)
	fmt.Printf("📊 API docs: http://localhost:%s/api/v1/stats\n", port)

	if err := app.Listen(":" + port); err != nil {
		log.Fatalf("❌ Failed to start server: %v", err)
	}
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
