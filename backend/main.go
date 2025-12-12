package main

import (
	"blackjack-api/game"
	"blackjack-api/handlers"
	"os"

	"github.com/gin-gonic/gin"
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

	// Dependencies
	playerStore := game.NewPlayerStore()
	gameController := handlers.NewGameController(playerStore)
	playerController := handlers.NewPlayerController(playerStore)

	api := r.Group("/api")
	{
		// Games
		api.POST("/games", gameController.StartGame)
		api.POST("/games/:id/action", gameController.PerformAction)

		// Players
		api.POST("/players", playerController.CreatePlayer)
		api.GET("/players/:id", playerController.GetPlayer)
		api.POST("/players/:id/reset", playerController.ResetBalance)
	}

	r.Run(":8080")
}
