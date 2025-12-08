package main

import (
	"github.com/gin-gonic/gin"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"mini-shodan-backend/internal/models"
    "mini-shodan-backend/internal/scanner"
)

func main() {
	// Connexion DB
	db, _ := gorm.Open(sqlite.Open("mini-shodan.db"), &gorm.Config{})
	
    // Migration automatique (Crée les tables si elles n'existent pas)
	db.AutoMigrate(&models.Host{}, &models.Port{})

	//Setup du Serveur Web
	r := gin.Default()

	// lancer un scan
	r.POST("/scan", func(c *gin.Context) {
        var json struct { Target string `json:"target"` }
        if err := c.ShouldBindJSON(&json); err != nil {
            c.JSON(400, gin.H{"error": "Target requise"})
            return
        }

 
        go func(ip string) {
            host, _ := scanner.ScanTarget(ip)
            if host != nil {
                db.Create(host) 
            }
        }(json.Target)

		c.JSON(200, gin.H{"message": "Scan démarré en arrière-plan", "target": json.Target})
	})

   
    r.GET("/hosts", func(c *gin.Context) {
        var hosts []models.Host
       
        db.Preload("Ports").Find(&hosts)
        c.JSON(200, hosts)
    })

	r.Run(":8080")
}