package game

// CalculateScore calculates the score of a hand according to Blackjack rules
func CalculateScore(hand []Card) int {
	score := 0
	aces := 0

	for _, card := range hand {
		value := 0
		switch card.Rank {
		case Two:
			value = 2
		case Three:
			value = 3
		case Four:
			value = 4
		case Five:
			value = 5
		case Six:
			value = 6
		case Seven:
			value = 7
		case Eight:
			value = 8
		case Nine:
			value = 9
		case Ten, Jack, Queen, King:
			value = 10
		case Ace:
			value = 11
			aces++
		}
		score += value
	}

	// Adjust for Aces
	for score > 21 && aces > 0 {
		score -= 10
		aces--
	}

	return score
}

// IsBust checks if a score is over 21
func IsBust(score int) bool {
	return score > 21
}

// ShouldDealerHit determines if the dealer should hit based on Soft 17 rule
// Dealer MUST HIT on Soft 17 (Ace + 6 treated as 17) or any total < 17.
// Dealer MUST STAND on Hard 17 or higher.
func ShouldDealerHit(hand Hand) bool {
	score := hand.Score
	if score < 17 {
		return true
	}

	// Check for Soft 17: Score is 17 and there is an Ace that is being counted as 11.
	// If score is 17, let's see if we have an ace.
	if score == 17 {
		// A hand is "soft" if it has an Ace counted as 11.
		// Recalculate without ace adjustment logic to see 'raw' value?
		// Or simpler: If we have an Ace, and treating it as 1 makes the score <= 6 (impossible for 17 total)
		// Wait, Soft 17 means Ace + 6.
		// If we have an Ace, and the score is 17, it's Soft 17 UNLESS the Ace is forced to be 1.
		// Example: Ace, 6 -> 17 (Soft). Hit.
		// Example: Ace, 5, Ace -> 17 (Soft). Hit.
		// Example: 10, 6, Ace -> 17 (Hard, Ace is 1). Stand.

		hasAce := false
		for _, card := range hand.Cards {
			if card.Rank == Ace {
				hasAce = true
				break
			}
		}

		if !hasAce {
			return false // Hard 17, no ace.
		}

		// If we have an ace, is it soft?
		// Calculate score treating all Aces as 1.
		minScore := 0
		for _, card := range hand.Cards {
			val := 0
			switch card.Rank {
			case Two:
				val = 2
			case Three:
				val = 3
			case Four:
				val = 4
			case Five:
				val = 5
			case Six:
				val = 6
			case Seven:
				val = 7
			case Eight:
				val = 8
			case Nine:
				val = 9
			case Ten, Jack, Queen, King:
				val = 10
			case Ace:
				val = 1
			}
			minScore += val
		}

		// If minScore (aces as 1) is different from current Score (which is 17),
		// then an Ace is being counted as 11.
		// 17 - 10 = 7. If minScore is <= 7, then we have room to count an Ace as 11.
		// Actually, if score is 17, and minScore is 7, then yes, one Ace is 11. 7 + 10 = 17.
		// If minScore is 17, then no Ace is 11.

		if minScore != score {
			return true // It's Soft 17
		}
	}

	return false
}

// DealCard pops a card from the deck
func DealCard(deck *[]Card) Card {
	if len(*deck) == 0 {
		// Should handle empty deck, maybe reshuffle?
		// For this scope, let's assume deck is sufficient for a single game or simple error.
		return Card{}
	}
	card := (*deck)[0]
	*deck = (*deck)[1:]
	return card
}
