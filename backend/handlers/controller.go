package handlers

import (
	"blackjack-api/game"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type GameController struct {
	Store       *game.GameStore
	PlayerStore *game.PlayerStore
}

func NewGameController(playerStore *game.PlayerStore) *GameController {
	return &GameController{
		Store:       game.NewGameStore(),
		PlayerStore: playerStore,
	}
}

// GameResponse DTO to hide internal details if needed (e.g., hidden dealer card)
type GameResponse struct {
	ID            string          `json:"id"`
	PlayerID      string          `json:"player_id"`
	PlayerHand    game.Hand       `json:"player_hand"`
	DealerHand    game.Hand       `json:"dealer_hand"` // We might need to mask this
	Status        game.GameStatus `json:"status"`
	CurrentBet    int             `json:"current_bet"`
	Payout        float64         `json:"payout"`
	PlayerBalance int             `json:"player_balance"` // Current Balance
}

// StartGameRequest DTO
type StartGameRequest struct {
	PlayerID  string `json:"player_id" binding:"required"`
	BetAmount int    `json:"bet_amount" binding:"required,min=1,max=10"`
}

// StartGame handles POST /api/games
func (c *GameController) StartGame(ctx *gin.Context) {
	var req StartGameRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// Validate Player and Funds
	_, err := c.PlayerStore.AdjustBalance(req.PlayerID, -req.BetAmount)
	if err != nil {
		// e.g. "player not found" or "insufficient funds"
		status := http.StatusBadRequest
		if err.Error() == "player not found" {
			status = http.StatusNotFound
		}
		ctx.JSON(status, gin.H{"error": err.Error()})
		return
	}

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
		PlayerID:   req.PlayerID,
		PlayerHand: playerHand,
		DealerHand: dealerHand,
		Deck:       deck,
		Status:     game.StatusPlayerTurn,
		CurrentBet: req.BetAmount,
		Payout:     0,
	}

	// Check for initial Blackjack
	payoutMultiplier := 0.0
	finished := false

	if playerHand.Score == 21 {
		if dealerHand.Score == 21 {
			gameState.Status = game.StatusPush
			payoutMultiplier = 1.0 // Return bet
		} else {
			gameState.Status = game.StatusPlayerWon
			payoutMultiplier = 2.5 // Blackjack pays 3:2 (Bet * 1 + Bet * 1.5) = Bet * 2.5
		}
		finished = true
	} else if dealerHand.Score == 21 {
		// Dealer blackjack, player loses (unless push handled above)
		gameState.Status = game.StatusDealerWon
		payoutMultiplier = 0.0
		finished = true
	}

	if finished {
		gameState.Payout = float64(gameState.CurrentBet) * payoutMultiplier
		if gameState.Payout > 0 {
			// Credit winnings
			// Note: We already deducted the bet. So if we "return bet", we credit bet amount.
			// If we win 3:2, we credit bet * 2.5.
			c.PlayerStore.AdjustBalance(gameState.PlayerID, int(gameState.Payout))
		}
	}

	c.Store.Save(gameState)

	// Get fresh balance for response
	player, _ := c.PlayerStore.Get(req.PlayerID)
	balance := 0
	if player != nil {
		balance = player.Balance
	}

	resp := maskDealerHand(gameState)
	resp.PlayerBalance = balance

	ctx.JSON(http.StatusCreated, resp)
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
			gameState.Payout = 0
		}

		c.Store.Save(gameState)

		player, _ := c.PlayerStore.Get(gameState.PlayerID)
		balance := 0
		if player != nil {
			balance = player.Balance
		}

		resp := maskDealerHand(gameState)
		resp.PlayerBalance = balance

		ctx.JSON(http.StatusOK, resp)
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
		payoutMultiplier := 0.0
		if game.IsBust(gameState.DealerHand.Score) {
			gameState.Status = game.StatusPlayerWon
			payoutMultiplier = 2.0 // 1:1 payout (Bet + Winnings)
		} else if gameState.DealerHand.Score > gameState.PlayerHand.Score {
			gameState.Status = game.StatusDealerWon
			payoutMultiplier = 0.0
		} else if gameState.DealerHand.Score < gameState.PlayerHand.Score {
			gameState.Status = game.StatusPlayerWon
			payoutMultiplier = 2.0
		} else {
			gameState.Status = game.StatusPush
			payoutMultiplier = 1.0 // Return bet
		}

		gameState.Payout = float64(gameState.CurrentBet) * payoutMultiplier
		if gameState.Payout > 0 {
			c.PlayerStore.AdjustBalance(gameState.PlayerID, int(gameState.Payout))
		}

		c.Store.Save(gameState)

		// Return full state (reveal dealer hand)
		player, _ := c.PlayerStore.Get(gameState.PlayerID)
		balance := 0
		if player != nil {
			balance = player.Balance
		}

		resp := GameResponse{
			ID:            gameState.ID,
			PlayerID:      gameState.PlayerID,
			PlayerHand:    gameState.PlayerHand,
			DealerHand:    gameState.DealerHand,
			Status:        gameState.Status,
			CurrentBet:    gameState.CurrentBet,
			Payout:        gameState.Payout,
			PlayerBalance: balance,
		}

		ctx.JSON(http.StatusOK, resp)
		return

	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
		return
	}
}

// maskDealerHand hides the dealer's second card if the game is still in progress
func maskDealerHand(g *game.GameState) GameResponse {
	// Create a copy or a DTO
	maskedHand := g.DealerHand

	// If game is over, show everything
	if g.Status == game.StatusPlayerTurn || g.Status == game.StatusDealerTurn {
		if len(maskedHand.Cards) > 1 {
			// Keep only the first card visible
			maskedHand.Cards = []game.Card{maskedHand.Cards[0], {Rank: "", Suit: ""}} // Mask second card
			// Don't show score or show partial? Usually hide score too.
			maskedHand.Score = 0 // Or just calculate score of first card
		}
	}

	return GameResponse{
		ID:         g.ID,
		PlayerID:   g.PlayerID,
		PlayerHand: g.PlayerHand,
		DealerHand: maskedHand,
		Status:     g.Status,
		CurrentBet: g.CurrentBet,
		Payout:     g.Payout,
		// PlayerBalance must be filled by caller
	}
}
