package game

import (
	"sync"
)

// GameStore is a thread-safe in-memory store for games
type GameStore struct {
	mu    sync.RWMutex
	games map[string]*GameState
}

// NewGameStore creates a new GameStore
func NewGameStore() *GameStore {
	return &GameStore{
		games: make(map[string]*GameState),
	}
}

// Save stores a game state
func (s *GameStore) Save(game *GameState) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.games[game.ID] = game
}

// Get retrieves a game state by ID
func (s *GameStore) Get(id string) (*GameState, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	game, exists := s.games[id]
	return game, exists
}

// Delete removes a game from the store
func (s *GameStore) Delete(id string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.games, id)
}
