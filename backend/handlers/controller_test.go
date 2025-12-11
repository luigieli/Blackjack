package handlers

import (
	"blackjack-api/game"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
)

func setupTestRouter() *gin.Engine {
	gin.SetMode(gin.TestMode)
	r := gin.Default()
	playerStore := game.NewPlayerStore()
	gameController := NewGameController(playerStore)
	playerController := NewPlayerController(playerStore)

	api := r.Group("/api")
	{
		api.POST("/games", gameController.StartGame)
		api.POST("/games/:id/action", gameController.PerformAction)
		api.POST("/players", playerController.CreatePlayer)
		api.POST("/players/:id/reset", playerController.ResetBalance)
	}
	return r
}

func TestGameFlow(t *testing.T) {
	router := setupTestRouter()

	// 1. Create Player
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/players", nil)
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var player game.Player
	json.Unmarshal(w.Body.Bytes(), &player)
	assert.NotEmpty(t, player.ID)
	assert.Equal(t, 100, player.Balance)

	playerID := player.ID

	// 2. Start Game
	w = httptest.NewRecorder()
	reqBody := StartGameRequest{
		PlayerID:  playerID,
		BetAmount: 10,
	}
	body, _ := json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/games", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var gameResp GameResponse
	json.Unmarshal(w.Body.Bytes(), &gameResp)

	// Balance logic:
	// If natural blackjack (instant win), balance = 100 - 10 + 25 = 115
	// If push (instant), balance = 100 - 10 + 10 = 100
	// If dealer blackjack (instant loss), balance = 100 - 10 = 90
	// If continuing, balance = 100 - 10 = 90

	expectedBalance := 90
	if gameResp.Status == game.StatusPlayerWon {
		expectedBalance = 115
	} else if gameResp.Status == game.StatusPush {
		expectedBalance = 100
	} else if gameResp.Status == game.StatusDealerWon {
		expectedBalance = 90
	}

	assert.Equal(t, expectedBalance, gameResp.PlayerBalance)
	assert.Equal(t, 10, gameResp.CurrentBet)

	gameID := gameResp.ID

	// If game is over immediately (e.g. natural blackjack), skip action test
	if gameResp.Status != game.StatusPlayerTurn {
		return
	}

	// 3. Stand (to finish game)
	w = httptest.NewRecorder()
	actionReq := ActionRequest{Action: "stand"}
	body, _ = json.Marshal(actionReq)
	req, _ = http.NewRequest("POST", "/api/games/"+gameID+"/action", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	router.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var finalResp GameResponse
	json.Unmarshal(w.Body.Bytes(), &finalResp)

	// Balance should be updated based on result
	if finalResp.Status == game.StatusPlayerWon {
		// Should be > 90. If 1:1, 90 + 20 = 110.
		assert.Greater(t, finalResp.PlayerBalance, 90)
	} else if finalResp.Status == game.StatusPush {
		assert.Equal(t, 100, finalResp.PlayerBalance)
	} else {
		assert.Equal(t, 90, finalResp.PlayerBalance)
	}
}

func TestInsufficientFunds(t *testing.T) {
	// Since we can't easily inject state into the router's closure without exposing it,
	// we will create a player, bet 100 to drain funds (or just bet more than balance).

	router := setupTestRouter()

	// 1. Create Player
	w := httptest.NewRecorder()
	req, _ := http.NewRequest("POST", "/api/players", nil)
	router.ServeHTTP(w, req)
	var player game.Player
	json.Unmarshal(w.Body.Bytes(), &player)

	// 2. Bet 200 (Balance 100)
	w = httptest.NewRecorder()
	reqBody := StartGameRequest{
		PlayerID:  player.ID,
		BetAmount: 10, // Max bet is 10, so we can't drain it quickly unless we loop.
		// Wait, user requirement says max bet 10.
		// So to test insufficient funds, we need to drain it or hack the store?
		// Since store is hidden, we can loop.
	}

	// Let's try to bet 100 (which is invalid due to max 10 validation).
	// So first we test validation.
	reqBodyInvalid := StartGameRequest{
		PlayerID:  player.ID,
		BetAmount: 100,
	}
	body, _ := json.Marshal(reqBodyInvalid)
	req, _ = http.NewRequest("POST", "/api/games", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	w = httptest.NewRecorder()
	router.ServeHTTP(w, req)
	assert.Equal(t, http.StatusBadRequest, w.Code)

	// Now let's drain funds?
	// Or we can just modify the test setup to expose the store.
	// But `setupTestRouter` creates new stores.
	// Let's copy paste `setupTestRouter` content here to access store.

	playerStore := game.NewPlayerStore()
	gameController := NewGameController(playerStore)

	// Create broke player
	p := playerStore.Create()
	playerStore.AdjustBalance(p.ID, -100)

	// Use controller directly for this specific test case to avoid router boilerplate if preferred,
	// OR construct router with this store.
	r := gin.Default()
	r.POST("/api/games", gameController.StartGame)

	w = httptest.NewRecorder()
	reqBody = StartGameRequest{
		PlayerID:  p.ID,
		BetAmount: 10,
	}
	body, _ = json.Marshal(reqBody)
	req, _ = http.NewRequest("POST", "/api/games", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	// Expect error
	assert.Equal(t, http.StatusBadRequest, w.Code)
	// Body should say "insufficient funds"
	assert.Contains(t, w.Body.String(), "insufficient funds")
}
