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
	ID               string          `json:"id"`
	PlayerHand       game.Hand       `json:"player_hand"`
	SplitHand        *game.Hand      `json:"split_hand,omitempty"`
	CurrentHandIndex int             `json:"current_hand_index"`
	DealerHand       game.Hand       `json:"dealer_hand"` // We might need to mask this
	Status           game.GameStatus `json:"status"`
	PlayerBalance    int             `json:"player_balance"`
	CurrentBet       int             `json:"current_bet"`
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

	if req.BetAmount < 1 {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Bet amount must be at least 1"})
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
	Action string `json:"action" binding:"required"` // "hit" or "stand" or "split"
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
	if !pExists {
		player = &game.Player{ID: gameState.PlayerID, Balance: 0}
	}

	if gameState.Status != game.StatusPlayerTurn {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Game is already over or not player's turn"})
		return
	}

	if req.Action == "split" {
		// Validations
		// 1. Can split only if not already split (simple version)
		if gameState.SplitHand != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Cannot split again"})
			return
		}
		// 2. Can split only if 2 cards in hand
		if len(gameState.PlayerHand.Cards) != 2 {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Can only split with 2 cards"})
			return
		}
		// 3. Can split only if ranks match
		if gameState.PlayerHand.Cards[0].Rank != gameState.PlayerHand.Cards[1].Rank {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Can only split cards of same rank"})
			return
		}
		// 4. Check balance
		if player.Balance < gameState.BetAmount {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": "Insufficient funds to split"})
			return
		}

		// Perform Split
		player.Balance -= gameState.BetAmount
		c.PlayerStore.Save(player)

		card2 := gameState.PlayerHand.Cards[1]

		// Setup First Hand
		gameState.PlayerHand.Cards = []game.Card{gameState.PlayerHand.Cards[0]}
		// Deal 2nd card to first hand
		gameState.PlayerHand.Cards = append(gameState.PlayerHand.Cards, game.DealCard(&gameState.Deck))
		gameState.PlayerHand.Score = game.CalculateScore(gameState.PlayerHand.Cards)

		// Setup Split Hand
		gameState.SplitHand = &game.Hand{
			Cards: []game.Card{card2},
		}
		// Deal 2nd card to split hand
		gameState.SplitHand.Cards = append(gameState.SplitHand.Cards, game.DealCard(&gameState.Deck))
		gameState.SplitHand.Score = game.CalculateScore(gameState.SplitHand.Cards)

		gameState.CurrentHandIndex = 0 // Start with first hand

		// Important: In standard Blackjack, if you split Aces, you get 1 card each and stand automatically.
		// Simplifying: Play normally for now unless specifically asked otherwise.

		c.Store.Save(gameState)
		ctx.JSON(http.StatusOK, c.maskDealerHand(gameState, player.Balance))
		return
	}

	if req.Action == "hit" {
		// Determine which hand to hit
		var activeHand *game.Hand
		if gameState.CurrentHandIndex == 0 {
			activeHand = &gameState.PlayerHand
		} else if gameState.CurrentHandIndex == 1 && gameState.SplitHand != nil {
			activeHand = gameState.SplitHand
		} else {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Invalid hand state"})
			return
		}

		card := game.DealCard(&gameState.Deck)
		activeHand.Cards = append(activeHand.Cards, card)
		activeHand.Score = game.CalculateScore(activeHand.Cards)

		if game.IsBust(activeHand.Score) {
			// If split, move to next hand?
			if gameState.SplitHand != nil && gameState.CurrentHandIndex == 0 {
				gameState.CurrentHandIndex = 1
			} else {
				// Game Over (unless Dealer needs to play for other hand? No, if bust here, we settle this hand)
				// Wait. Logic for Split:
				// If Hand 1 busts, it loses immediately. Then play Hand 2.
				// If Hand 2 busts, it loses immediately.
				// Dealer plays if AT LEAST one hand did not bust.

				if gameState.SplitHand != nil && gameState.CurrentHandIndex == 1 {
					// Both hands played. Now Dealer?
					// Only if one of them is valid (not bust).
					// If Hand 1 Bust AND Hand 2 Bust -> Dealer doesn't play. Status DealerWon.

					// Let's defer "Dealer Turn" logic to a common finish step.
					// For now, if active hand busts, we treat it as "Stand" effectively but marked as bust.
					// But we need to know it busted.
					// Let's transition:
					c.finishPlayerTurn(gameState)
				} else if gameState.SplitHand == nil {
					// Single hand bust -> Game Over
					gameState.Status = game.StatusDealerWon
				}
			}
		}

		c.Store.Save(gameState)
		ctx.JSON(http.StatusOK, c.maskDealerHand(gameState, player.Balance))
		return

	} else if req.Action == "stand" {
		// If split and on first hand, move to second
		if gameState.SplitHand != nil && gameState.CurrentHandIndex == 0 {
			gameState.CurrentHandIndex = 1
			c.Store.Save(gameState)
			ctx.JSON(http.StatusOK, c.maskDealerHand(gameState, player.Balance))
			return
		}

		// Otherwise finish player turn
		c.finishPlayerTurn(gameState)

		// Update balance logic is complex with split.
		// We need to calculate winnings for each hand against dealer.
		// c.finishPlayerTurn handles dealer play. Now we calculate winnings.

		totalWinnings := 0

		// If dealer played (or we are resolving), compare hands.
		// Note: finishPlayerTurn sets status.

		// Helper to compare one hand
		resolveHand := func(hand game.Hand, bet int) int {
			if game.IsBust(hand.Score) {
				return 0 // Lost
			}
			if game.IsBust(gameState.DealerHand.Score) {
				return bet * 2
			}
			if hand.Score > gameState.DealerHand.Score {
				return bet * 2
			}
			if hand.Score == gameState.DealerHand.Score {
				return bet // Push
			}
			return 0
		}

		// Calculate winnings
		// Check Hand 1
		totalWinnings += resolveHand(gameState.PlayerHand, gameState.BetAmount)

		// Check Hand 2
		if gameState.SplitHand != nil {
			totalWinnings += resolveHand(*gameState.SplitHand, gameState.BetAmount)
		}

		// Update Balance if any winnings
		if totalWinnings > 0 {
			player.Balance += totalWinnings
			c.PlayerStore.Save(player)
		}

		// Set final Status for UI
		// If split, status "PlayerWon" or "DealerWon" is ambiguous.
		// Maybe just "GameOver"? Or keep it simple.
		// If any hand won, maybe "PlayerWon"?
		// Actually the UI usually shows result per hand or just balance update.
		// Let's set StatusDealerTurn -> Status... something.
		// Actually, finishPlayerTurn sets it to something?
		// We need to decide what 'Status' means now.
		// Let's say if all hands bust, DealerWon.
		// If at least one hand wins, PlayerWon?
		// Let's stick to: If game over, status is not PlayerTurn.

		c.Store.Save(gameState)
		ctx.JSON(http.StatusOK, c.maskDealerHand(gameState, player.Balance))
		return

	} else {
		ctx.JSON(http.StatusBadRequest, gin.H{"error": "Invalid action"})
		return
	}
}

func (c *GameController) finishPlayerTurn(gameState *game.GameState) {
	// Check if all player hands are busted.
	allBusted := true
	if !game.IsBust(gameState.PlayerHand.Score) {
		allBusted = false
	}
	if gameState.SplitHand != nil && !game.IsBust(gameState.SplitHand.Score) {
		allBusted = false
	}

	if allBusted {
		gameState.Status = game.StatusDealerWon // Or simple "Game Over"
		return
	}

	gameState.Status = game.StatusDealerTurn

	// Dealer plays
	for game.ShouldDealerHit(gameState.DealerHand) {
		card := game.DealCard(&gameState.Deck)
		gameState.DealerHand.Cards = append(gameState.DealerHand.Cards, card)
		gameState.DealerHand.Score = game.CalculateScore(gameState.DealerHand.Cards)
	}

	// Determine generic status (mostly for UI color)
	if game.IsBust(gameState.DealerHand.Score) {
		gameState.Status = game.StatusPlayerWon
	} else {
		// Complex result. Let's just say "finished" implicitly by not being PlayerTurn.
		// We can reuse PlayerWon/DealerWon/Push if single hand.
		if gameState.SplitHand == nil {
			if gameState.DealerHand.Score > gameState.PlayerHand.Score {
				gameState.Status = game.StatusDealerWon
			} else if gameState.DealerHand.Score < gameState.PlayerHand.Score {
				gameState.Status = game.StatusPlayerWon
			} else {
				gameState.Status = game.StatusPush
			}
		} else {
			// Split results are mixed. Just mark as "DealerWon" (Game Over)
			// and let UI show scores? Or add a Status "GameOver".
			// Since we use strings, let's add "GameOver" or just leave as DealerTurn?
			// No, DealerTurn implies waiting.
			// Let's use "PlayerWon" if net profit > 0 ?
			// Let's use "Push" as generic "Game Over" for split?
			// User didn't specify. I'll stick to a generic status or just pick one.
			// Let's leave it as "DealerWon" (meaning round finished)
			// but relies on resolve logic to pay.
			// Actually, "DealerWon" shows Red. "PlayerWon" shows Green.
			// Maybe check net result?
			// For now, let's use "PlayerWon" if at least one hand won?
			gameState.Status = game.StatusPush // Neutral color?
		}
	}
}

// maskDealerHand hides the dealer's second card if the game is still in progress
func (c *GameController) maskDealerHand(g *game.GameState, balance int) GameResponse {
	// If game is over, show everything
	if g.Status != game.StatusPlayerTurn && g.Status != game.StatusDealerTurn {
		return GameResponse{
			ID:               g.ID,
			PlayerHand:       g.PlayerHand,
			SplitHand:        g.SplitHand,
			CurrentHandIndex: g.CurrentHandIndex,
			DealerHand:       g.DealerHand,
			Status:           g.Status,
			PlayerBalance:    balance,
			CurrentBet:       g.BetAmount,
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
		ID:               g.ID,
		PlayerHand:       g.PlayerHand,
		SplitHand:        g.SplitHand,
		CurrentHandIndex: g.CurrentHandIndex,
		DealerHand:       maskedHand,
		Status:           g.Status,
		PlayerBalance:    balance,
		CurrentBet:       g.BetAmount,
	}
}
