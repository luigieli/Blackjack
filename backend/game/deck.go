package game

import (
	"math/rand"
	"time"
)

var rng *rand.Rand

func init() {
	source := rand.NewSource(time.Now().UnixNano())
	rng = rand.New(source)
}

// NewDeck creates a standard 52-card deck
func NewDeck() []Card {
	suits := []Suit{Hearts, Diamonds, Clubs, Spades}
	ranks := []Rank{Two, Three, Four, Five, Six, Seven, Eight, Nine, Ten, Jack, Queen, King, Ace}

	var deck []Card
	for _, suit := range suits {
		for _, rank := range ranks {
			deck = append(deck, Card{Suit: suit, Rank: rank})
		}
	}
	return deck
}

// Shuffle randomizes the order of cards in the deck
func Shuffle(deck []Card) []Card {
	shuffled := make([]Card, len(deck))
	copy(shuffled, deck)
	rng.Shuffle(len(shuffled), func(i, j int) {
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	})
	return shuffled
}
