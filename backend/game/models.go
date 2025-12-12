package game

// Suit represents the suit of a card
type Suit string

const (
	Hearts   Suit = "Hearts"
	Diamonds Suit = "Diamonds"
	Clubs    Suit = "Clubs"
	Spades   Suit = "Spades"
)

// Rank represents the rank of a card
type Rank string

const (
	Two   Rank = "2"
	Three Rank = "3"
	Four  Rank = "4"
	Five  Rank = "5"
	Six   Rank = "6"
	Seven Rank = "7"
	Eight Rank = "8"
	Nine  Rank = "9"
	Ten   Rank = "10"
	Jack  Rank = "J"
	Queen Rank = "Q"
	King  Rank = "K"
	Ace   Rank = "A"
)

// Card represents a single playing card
type Card struct {
	Suit  Suit `json:"suit"`
	Rank  Rank `json:"rank"`
	Value int  `json:"value"` // We can store the base value (2-10, 11 for Ace initially?) or calculate it.
	// Actually, context says 2-9 face, 10 JQK=10, Ace=1/11 dynamic.
	// Storing a static value might be misleading for Ace.
	// Let's store a "default" value (e.g. 11 for Ace) or 0 and calculate later.
	// Let's omit Value field from JSON for now and calculate in Engine,
	// OR Keep it if the frontend needs it.
	// Let's keep it simple: Rank and Suit are the identity.
}

// Hand represents a player's or dealer's hand
type Hand struct {
	Cards []Card `json:"cards"`
	Score int    `json:"score"` // Calculated score
}

// GameStatus represents the current state of the game
type GameStatus string

const (
	StatusPlayerTurn GameStatus = "PlayerTurn"
	StatusDealerTurn GameStatus = "DealerTurn"
	StatusPlayerWon  GameStatus = "PlayerWon"
	StatusDealerWon  GameStatus = "DealerWon" // Dealer wins or Player busts
	StatusPush       GameStatus = "Push"
)

// GameState represents the entire state of a blackjack game
type GameState struct {
	ID         string     `json:"id"`
	PlayerID   string     `json:"player_id"`
	BetAmount  int        `json:"bet_amount"`
	PlayerHand Hand       `json:"player_hand"`
	DealerHand Hand       `json:"dealer_hand"`
	Deck       []Card     `json:"-"` // Hide deck from JSON
	Status     GameStatus `json:"status"`
}
