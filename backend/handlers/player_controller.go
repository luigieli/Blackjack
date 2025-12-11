package handlers

import (
	"blackjack-api/game"
	"net/http"

	"github.com/gin-gonic/gin"
)

type PlayerController struct {
	Store *game.PlayerStore
}

func NewPlayerController(store *game.PlayerStore) *PlayerController {
	return &PlayerController{
		Store: store,
	}
}

// CreatePlayer handles POST /api/players
func (c *PlayerController) CreatePlayer(ctx *gin.Context) {
	player := c.Store.Create()
	ctx.JSON(http.StatusCreated, player)
}

// ResetBalance handles POST /api/players/:id/reset
func (c *PlayerController) ResetBalance(ctx *gin.Context) {
	id := ctx.Param("id")
	player, err := c.Store.ResetBalance(id)
	if err != nil {
		ctx.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, player)
}

// GetPlayer handles GET /api/players/:id
func (c *PlayerController) GetPlayer(ctx *gin.Context) {
	id := ctx.Param("id")
	player, exists := c.Store.Get(id)
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Player not found"})
		return
	}
	ctx.JSON(http.StatusOK, player)
}
