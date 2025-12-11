package main

import (
	"blackjack-api/game"
	"blackjack-api/handlers"
	"log"
	"os"

	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Determine frontend path
	frontendPath := "./frontend"
	if _, err := os.Stat(frontendPath); os.IsNotExist(err) {
		if _, err := os.Stat("../frontend"); err == nil {
			frontendPath = "../frontend"
		} else {
			log.Printf("Warning: frontend directory not found at ./frontend or ../frontend")
		}
	}
	log.Printf("Serving frontend from: %s", frontendPath)

	// Static files
	r.Static("/frontend", frontendPath)
	r.StaticFile("/", frontendPath+"/index.html")
	r.StaticFile("/app.js", frontendPath+"/app.js")

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
