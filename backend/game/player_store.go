package game

import (
	"sync"
)

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

// Save stores a player state
func (s *PlayerStore) Save(player *Player) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.players[player.ID] = player
}

// Get retrieves a player by ID
func (s *PlayerStore) Get(id string) (*Player, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	player, exists := s.players[id]
	return player, exists
}
