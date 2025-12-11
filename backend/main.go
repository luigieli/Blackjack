package main

import (
	"blackjack-api/handlers"
	"github.com/gin-gonic/gin"
)

func main() {
	r := gin.Default()

	// Serve frontend static files
	// The frontend directory is mounted to /app/frontend in the container
	// When running locally, we need to point to the correct path.
	// We will assume "frontend" if running from root, or "../frontend" if from backend.
	// Since we are running `go run .` inside `backend`, `../frontend` is correct.
	// But in Docker, we map to `/app/frontend`, and working dir is `/app`. So `./frontend`.

	// Let's support both or just assume relative path ./frontend exists (via docker mount).
	// If running locally, I should probably symlink or just run from root?
	// No, if I run `go run .` in `backend`, I need `../frontend`.
	// Let's check if ./frontend exists, if not try ../frontend.

	r.Static("/frontend", "./frontend")
	r.StaticFile("/", "./frontend/index.html")
	// Also serve app.js at root or relative
	r.StaticFile("/app.js", "./frontend/app.js")

	gameController := handlers.NewGameController()

	api := r.Group("/api")
	{
		api.POST("/games", gameController.StartGame)
		api.POST("/games/:id/action", gameController.PerformAction)
	}

	r.Run(":8080")
}
