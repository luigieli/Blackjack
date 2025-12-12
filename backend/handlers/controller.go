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

func NewGameController() *GameController {
	return &GameController{
		Store:       game.NewGameStore(),
		PlayerStore: game.NewPlayerStore(),
	}
}

// GameResponse DTO to hide internal details if needed (e.g., hidden dealer card)
type GameResponse struct {
	ID            string          `json:"id"`
	PlayerHand    game.Hand       `json:"player_hand"`
	DealerHand    game.Hand       `json:"dealer_hand"` // We might need to mask this
	Status        game.GameStatus `json:"status"`
	PlayerBalance int             `json:"player_balance"`
	CurrentBet    int             `json:"current_bet"`
}

type StartGameRequest struct {
	BetAmount int `json:"bet_amount" binding:"required"`
}

// StartGame handles POST /api/games
func (c *GameController) StartGame(ctx *gin.Context) {
	playerID := ctx.GetHeader("X-Player-ID")
	if playerID == "" {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "X-Player-ID header is required"})
		return
	}

	var req StartGameRequest
	if err := ctx.ShouldBindJSON(&req); err != nil {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "bet_amount is required and must be an integer"})
		return
	}

	if req.BetAmount < 1 || req.BetAmount > 10 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Bet amount must be between 1 and 10"})
		return
	}

	// Get or Create Player
	player, exists := c.PlayerStore.Get(playerID)
	if !exists {
		player = &game.Player{ID: playerID, Balance: 100}
		c.PlayerStore.Save(player)
	}

	// Auto-top-up if balance is 0 (Bankrupt)
	if player.Balance <= 0 {
		player.Balance = 100
		c.PlayerStore.Save(player)
	}

	// Validate Balance
	if player.Balance < req.BetAmount {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds"})
		return
	}

	// Deduct Bet
	player.Balance -= req.BetAmount
	c.PlayerStore.Save(player)

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
		PlayerID:   playerID,
		BetAmount:  req.BetAmount,
		PlayerHand: playerHand,
		DealerHand: dealerHand,
		Deck:       deck,
		Status:     game.StatusPlayerTurn,
	}

	// Check for initial Blackjack
	if playerHand.Score == 21 {
		if dealerHand.Score == 21 {
			gameState.Status = game.StatusPush
			// Refund Bet
			player.Balance += req.BetAmount
		} else {
			gameState.Status = game.StatusPlayerWon
			// Blackjack Payout (3:2) -> Return Bet + 1.5 * Bet = 2.5 * Bet
			// Since we already deducted the bet, we add 2.5 * Bet back.
			// E.g. Bet 10. Balance -10. Win. Balance += 25. Net +15.
			payout := int(float64(req.BetAmount) * 2.5)
			player.Balance += payout
		}
		c.PlayerStore.Save(player)
	} else if dealerHand.Score == 21 {
		// Dealer blackjack, player loses (unless push handled above)
		gameState.Status = game.StatusDealerWon
		// No refund
	}

	c.Store.Save(gameState)

	ctx.JSON(http.StatusCreated, c.maskDealerHand(gameState, player.Balance))
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

	// Fetch Player for balance updates
	player, pExists := c.PlayerStore.Get(gameState.PlayerID)
	// If player is missing for some reason, re-create or error?
	// This shouldn't happen in normal flow, but let's handle gracefully.
	if !pExists {
		// Just re-create to avoid crash, but they lost their balance history.
		player = &game.Player{ID: gameState.PlayerID, Balance: 0}
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
			// Player loses bet (already deducted)
		}

		c.Store.Save(gameState)
		ctx.JSON(http.StatusOK, c.maskDealerHand(gameState, player.Balance))
		return

	} else if req.Action == "stand" {
		gameState.Status = game.StatusDealerTurn

		// Dealer Logic
		for game.ShouldDealerHit(gameState.DealerHand) {
			card := game.DealCard(&gameState.Deck)
			gameState.DealerHand.Cards = append(gameState.DealerHand.Cards, card)
			gameState.DealerHand.Score = game.CalculateScore(gameState.DealerHand.Cards)
		}

		// Determine Winner & Payout
		winnings := 0
		if game.IsBust(gameState.DealerHand.Score) {
			gameState.Status = game.StatusPlayerWon
			winnings = gameState.BetAmount * 2
		} else if gameState.DealerHand.Score > gameState.PlayerHand.Score {
			gameState.Status = game.StatusDealerWon
			winnings = 0
		} else if gameState.DealerHand.Score < gameState.PlayerHand.Score {
			gameState.Status = game.StatusPlayerWon
			winnings = gameState.BetAmount * 2
		} else {
			gameState.Status = game.StatusPush
			winnings = gameState.BetAmount
		}

		// Update Balance
		if winnings > 0 {
			player.Balance += winnings
			c.PlayerStore.Save(player)
		}

		c.Store.Save(gameState)
		// Return full state (reveal dealer hand)
		resp := c.maskDealerHand(gameState, player.Balance)
		// Force reveal because maskDealerHand hides it if status is special,
		// but wait, maskDealerHand reveals if status != PlayerTurn/DealerTurn.
		// Stand sets status to Won/Lost/Push, so it should reveal.
		// EXCEPT if status is DealerTurn... wait.
		// logic: Stand -> DealerTurn -> Loop -> Final Status.
		// So when we return here, status is Final. maskDealerHand will reveal.
		ctx.JSON(http.StatusOK, resp)
		return

	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
		return
	}
}

// maskDealerHand hides the dealer's second card if the game is still in progress
func (c *GameController) maskDealerHand(g *game.GameState, balance int) GameResponse {
	// If game is over, show everything
	if g.Status != game.StatusPlayerTurn && g.Status != game.StatusDealerTurn {
		return GameResponse{
			ID:            g.ID,
			PlayerHand:    g.PlayerHand,
			DealerHand:    g.DealerHand,
			Status:        g.Status,
			PlayerBalance: balance,
			CurrentBet:    g.BetAmount,
		}
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
		ID:            g.ID,
		PlayerHand:    g.PlayerHand,
		DealerHand:    maskedHand,
		Status:        g.Status,
		PlayerBalance: balance,
		CurrentBet:    g.BetAmount,
	}
}
