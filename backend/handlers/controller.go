package handlers

import (
	"blackjack-api/game"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GameController struct {
	Store *game.GameStore
}

func NewGameController() *GameController {
	return &GameController{
		Store: game.NewGameStore(),
	}
}

// GameResponse DTO to hide internal details if needed (e.g., hidden dealer card)
type GameResponse struct {
	ID         string          `json:"id"`
	PlayerHand game.Hand       `json:"player_hand"`
	DealerHand game.Hand       `json:"dealer_hand"` // We might need to mask this
	Status     game.GameStatus `json:"status"`
}

// StartGame handles POST /api/games
func (c *GameController) StartGame(ctx *gin.Context) {
	// Initialize Deck
	deck := game.NewDeck()
	deck = game.Shuffle(deck)

	// Deal initial cards
	playerHand := game.Hand{Cards: []game.Card{}}
	dealerHand := game.Hand{Cards: []game.Card{}}

	// Player gets 2 cards
	playerHand.Cards = append(playerHand.Cards, game.DealCard(&deck))
	playerHand.Cards = append(playerHand.Cards, game.DealCard(&deck))

	// Dealer gets 2 cards
	dealerHand.Cards = append(dealerHand.Cards, game.DealCard(&deck))
	dealerHand.Cards = append(dealerHand.Cards, game.DealCard(&deck))

	// Calculate initial scores
	playerHand.Score = game.CalculateScore(playerHand.Cards)
	dealerHand.Score = game.CalculateScore(dealerHand.Cards)

	id := uuid.New().String()

	gameState := &game.GameState{
		ID:         id,
		PlayerHand: playerHand,
		DealerHand: dealerHand,
		Deck:       deck,
		Status:     game.StatusPlayerTurn,
	}

	// Check for initial Blackjack
	if playerHand.Score == 21 {
		if dealerHand.Score == 21 {
			gameState.Status = game.StatusPush
		} else {
			gameState.Status = game.StatusPlayerWon
		}
	} else if dealerHand.Score == 21 {
		// Dealer blackjack, player loses (unless push handled above)
		gameState.Status = game.StatusDealerWon
	}

	c.Store.Save(gameState)

	ctx.JSON(http.StatusCreated, maskDealerHand(gameState))
}

// ActionRequest DTO
type ActionRequest struct {
	Action string `json:"action" binding:"required"` // "hit" or "stand"
}

// PerformAction handles POST /api/games/:id/action
func (c *GameController) PerformAction(ctx *gin.Context) {
	id := ctx.Param("id")
	var req ActionRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	gameState, exists := c.Store.Get(id)
	if !exists {
		ctx.JSON(http.StatusNotFound, gin.H{"error": "Game not found"})
		return
	}

	if gameState.Status != game.StatusPlayerTurn {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Game is already over or not player's turn"})
		return
	}

	if req.Action == "hit" {
		// Deal card to player
		card := game.DealCard(&gameState.Deck)
		gameState.PlayerHand.Cards = append(gameState.PlayerHand.Cards, card)
		gameState.PlayerHand.Score = game.CalculateScore(gameState.PlayerHand.Cards)

		if game.IsBust(gameState.PlayerHand.Score) {
			gameState.Status = game.StatusDealerWon // Player Bust
		}

		c.Store.Save(gameState)
		ctx.JSON(http.StatusOK, maskDealerHand(gameState))
		return

	} else if req.Action == "stand" {
		gameState.Status = game.StatusDealerTurn

		// Dealer Logic
		for game.ShouldDealerHit(gameState.DealerHand) {
			card := game.DealCard(&gameState.Deck)
			gameState.DealerHand.Cards = append(gameState.DealerHand.Cards, card)
			gameState.DealerHand.Score = game.CalculateScore(gameState.DealerHand.Cards)
		}

		// Determine Winner
		if game.IsBust(gameState.DealerHand.Score) {
			gameState.Status = game.StatusPlayerWon
		} else if gameState.DealerHand.Score > gameState.PlayerHand.Score {
			gameState.Status = game.StatusDealerWon
		} else if gameState.DealerHand.Score < gameState.PlayerHand.Score {
			gameState.Status = game.StatusPlayerWon
		} else {
			gameState.Status = game.StatusPush
		}

		c.Store.Save(gameState)
		// Return full state (reveal dealer hand)
		ctx.JSON(http.StatusOK, gameState)
		return

	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
		return
	}
}

// maskDealerHand hides the dealer's second card if the game is still in progress
func maskDealerHand(g *game.GameState) interface{} {
	// If game is over, show everything
	if g.Status != game.StatusPlayerTurn && g.Status != game.StatusDealerTurn {
		return g
	}

	// Create a copy or a DTO
	maskedHand := g.DealerHand
	if len(maskedHand.Cards) > 1 {
		// Keep only the first card visible
		maskedHand.Cards = []game.Card{maskedHand.Cards[0], {Rank: "", Suit: ""}} // Mask second card
		// Don't show score or show partial? Usually hide score too.
		maskedHand.Score = 0 // Or just calculate score of first card
	}

	return GameResponse{
		ID:         g.ID,
		PlayerHand: g.PlayerHand,
		DealerHand: maskedHand,
		Status:     g.Status,
	}
}
