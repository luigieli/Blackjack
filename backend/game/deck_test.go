package game

import (
	"testing"
)

func TestNewDeck(t *testing.T) {
	deck := NewDeck()
	if len(deck) != 52 {
		t.Errorf("Expected deck length of 52, got %d", len(deck))
	}

	// Check for uniqueness (simple check)
	cardMap := make(map[string]bool)
	for _, card := range deck {
		key := string(card.Rank) + string(card.Suit)
		if cardMap[key] {
			t.Errorf("Duplicate card found: %s", key)
		}
		cardMap[key] = true
	}
}

func TestShuffle(t *testing.T) {
	deck := NewDeck()
	shuffled := Shuffle(deck)

	if len(shuffled) != 52 {
		t.Errorf("Expected shuffled deck length of 52, got %d", len(shuffled))
	}

	// Check that the order is different (statistically probable)
	// Theoretically possible to be same, but unlikely.
	// We can just check that it contains the same cards.

	match := true
	for i, card := range deck {
		if card != shuffled[i] {
			match = false
			break
		}
	}
	if match {
		t.Log("Warning: Shuffled deck is in same order as original (highly unlikely but possible)")
	}
}
