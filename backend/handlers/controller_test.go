package handlers

import (
	"blackjack-api/game"
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestBettingFlow(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	controller := NewGameController()
	router := gin.Default()
	router.POST("/api/games", controller.StartGame)
	router.POST("/api/games/:id/action", controller.PerformAction)

	// Test 1: Start Game with Bet
	playerID := "test-player-1"
	betAmount := 10
	startReq := StartGameRequest{BetAmount: betAmount}
	body, _ := json.Marshal(startReq)

	req, _ := http.NewRequest("POST", "/api/games", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Player-ID", playerID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusCreated {
		t.Fatalf("Expected StatusCreated, got %v", w.Code)
	}

	var gameResp GameResponse
	json.Unmarshal(w.Body.Bytes(), &gameResp)

	// Verify Balance Deducted (100 - 10 = 90)
	// Unless initial blackjack occurred...
	if gameResp.Status == game.StatusPlayerWon && gameResp.PlayerHand.Score == 21 {
		// Blackjack! Balance should be 100 - 10 + 25 = 115
		if gameResp.PlayerBalance != 115 {
			t.Errorf("Expected balance 115 after BJ, got %d", gameResp.PlayerBalance)
		}
	} else if gameResp.Status == game.StatusPush {
		// Push (Double BJ)
		if gameResp.PlayerBalance != 100 {
			t.Errorf("Expected balance 100 after Push, got %d", gameResp.PlayerBalance)
		}
	} else if gameResp.Status == game.StatusDealerWon {
		// Dealer BJ
		if gameResp.PlayerBalance != 90 {
			t.Errorf("Expected balance 90 after Dealer BJ, got %d", gameResp.PlayerBalance)
		}
	} else {
		// Normal game start
		if gameResp.PlayerBalance != 90 {
			t.Errorf("Expected balance 90 after bet, got %d", gameResp.PlayerBalance)
		}

		// Force a Stand to check payout logic (assuming no bust)
		actionReq := ActionRequest{Action: "stand"}
		actionBody, _ := json.Marshal(actionReq)
		req2, _ := http.NewRequest("POST", "/api/games/"+gameResp.ID+"/action", bytes.NewBuffer(actionBody))
		req2.Header.Set("Content-Type", "application/json")
		req2.Header.Set("X-Player-ID", playerID)
		w2 := httptest.NewRecorder()
		router.ServeHTTP(w2, req2)

		if w2.Code != http.StatusOK {
			t.Fatalf("Expected StatusOK on stand, got %v", w2.Code)
		}

		var standResp GameResponse
		json.Unmarshal(w2.Body.Bytes(), &standResp)

		// Verify Payout logic
		expectedBalance := 90
		if standResp.Status == game.StatusPlayerWon {
			expectedBalance += 20 // 10 * 2
		} else if standResp.Status == game.StatusPush {
			expectedBalance += 10 // 10
		}
		// Loss = +0

		if standResp.PlayerBalance != expectedBalance {
			t.Errorf("Expected balance %d, got %d", expectedBalance, standResp.PlayerBalance)
		}
	}
}

func TestInsufficientFunds(t *testing.T) {
	// Setup
	gin.SetMode(gin.TestMode)
	controller := NewGameController()
	router := gin.Default()
	router.POST("/api/games", controller.StartGame)

	playerID := "broke-player"
	// Set player balance to 5 manually
	player := &game.Player{ID: playerID, Balance: 5}
	controller.PlayerStore.Save(player)

	// Try to bet 10
	startReq := StartGameRequest{BetAmount: 10}
	body, _ := json.Marshal(startReq)

	req, _ := http.NewRequest("POST", "/api/games", bytes.NewBuffer(body))
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("X-Player-ID", playerID)
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusBadRequest {
		t.Errorf("Expected BadRequest for insufficient funds, got %v", w.Code)
	}
}
