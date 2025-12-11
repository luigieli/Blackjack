package main

import (
	"blackjack-api/handlers"
	"github.com/gin-gonic/gin"
	"os"
)

func main() {
	r := gin.Default()

	// Determine web directory path
	// If running in Docker or from root where ./web exists
	webDir := "./web"
	if _, err := os.Stat(webDir); os.IsNotExist(err) {
		// Try ../web (if running from backend/)
		if _, err := os.Stat("../web"); err == nil {
			webDir = "../web"
		}
	}

	r.Static("/web", webDir)
	r.StaticFile("/", webDir+"/index.html")
	r.StaticFile("/app.js", webDir+"/app.js")
	r.StaticFile("/style.css", webDir+"/style.css")

	gameController := handlers.NewGameController()

	api := r.Group("/api")
	{
		api.POST("/games", gameController.StartGame)
		api.POST("/games/:id/action", gameController.PerformAction)
	}

	r.Run(":8080")
}
