package game

import (
	"errors"
	"sync"

	"github.com/google/uuid"
)

// Player represents a user in the system
type Player struct {
	ID      string `json:"id"`
	Balance int    `json:"balance"`
}

// PlayerStore is a thread-safe in-memory store for players
type PlayerStore struct {
	mu      sync.RWMutex
	players map[string]*Player
}

// NewPlayerStore creates a new PlayerStore
func NewPlayerStore() *PlayerStore {
	return &PlayerStore{
		players: make(map[string]*Player),
	}
}

// Create creates a new player with default balance
func (s *PlayerStore) Create() *Player {
	s.mu.Lock()
	defer s.mu.Unlock()

	id := uuid.New().String()
	player := &Player{
		ID:      id,
		Balance: 100, // Default starting balance
	}
	s.players[id] = player
	return player
}

// Get retrieves a player by ID
func (s *PlayerStore) Get(id string) (*Player, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	// Return a copy to avoid unsafe concurrent modification?
	// Or return pointer? Given GameStore returns pointer, we'll return pointer
	// but strictly we should operate via Store methods for mutation if possible.
	// But to keep it consistent with GameStore, let's return pointer.
	// However, for balance updates, I'll add a helper method to ensure safety.
	p, exists := s.players[id]
	return p, exists
}

// ResetBalance resets a player's balance to 100
func (s *PlayerStore) ResetBalance(id string) (*Player, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, exists := s.players[id]
	if !exists {
		return nil, errors.New("player not found")
	}
	p.Balance = 100
	return p, nil
}

// AdjustBalance adds amount to balance (can be negative). Returns new balance.
// Returns error if insufficient funds (for negative adjustment).
func (s *PlayerStore) AdjustBalance(id string, amount int) (int, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	p, exists := s.players[id]
	if !exists {
		return 0, errors.New("player not found")
	}

	if p.Balance+amount < 0 {
		return p.Balance, errors.New("insufficient funds")
	}

	p.Balance += amount
	return p.Balance, nil
}
