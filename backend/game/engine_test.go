package game

import "testing"

func TestCalculateScore(t *testing.T) {
	tests := []struct {
		name     string
		cards    []Card
		expected int
	}{
		{"Two cards", []Card{{Rank: Ten}, {Rank: Five}}, 15},
		{"Blackjack", []Card{{Rank: Ace}, {Rank: King}}, 21},
		{"Ace as 1", []Card{{Rank: Ace}, {Rank: Nine}, {Rank: Eight}}, 18}, // 1+9+8=18
		{"Ace as 11", []Card{{Rank: Ace}, {Rank: Five}}, 16},
		{"Multiple Aces", []Card{{Rank: Ace}, {Rank: Ace}}, 12}, // 11+1=12
		{"Bust with Ace", []Card{{Rank: Ten}, {Rank: Ten}, {Rank: Ace}}, 21},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			score := CalculateScore(tt.cards)
			if score != tt.expected {
				t.Errorf("expected %d, got %d", tt.expected, score)
			}
		})
	}
}

func TestShouldDealerHit(t *testing.T) {
	tests := []struct {
		name     string
		cards    []Card
		score    int // Pre-calculated score passed to Hand
		expected bool
	}{
		{"Hard 16", []Card{{Rank: Ten}, {Rank: Six}}, 16, true},
		{"Hard 17", []Card{{Rank: Ten}, {Rank: Seven}}, 17, false},
		{"Soft 17 (A+6)", []Card{{Rank: Ace}, {Rank: Six}}, 17, true},
		{"Soft 17 (A+3+3)", []Card{{Rank: Ace}, {Rank: Three}, {Rank: Three}}, 17, true},
		{"Hard 17 (10+6+A)", []Card{{Rank: Ten}, {Rank: Six}, {Rank: Ace}}, 17, false},
		{"18", []Card{{Rank: Ten}, {Rank: Eight}}, 18, false},
		{"Soft 18", []Card{{Rank: Ace}, {Rank: Seven}}, 18, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			hand := Hand{Cards: tt.cards, Score: tt.score}
			if got := ShouldDealerHit(hand); got != tt.expected {
				t.Errorf("expected %v, got %v", tt.expected, got)
			}
		})
	}
}
