# Project Context: Go Blackjack REST API

## 1. Mission Statement
Build a robust, containerized REST API that acts as a Blackjack Dealer for a single player. The system must implement stateful game logic over a stateless HTTP protocol using in-memory persistence, serving a lightweight Vanilla JS frontend.

## 2. Technical Constraints & Stack
*   **Language:** Go (Golang) - *Latest Stable*
*   **Framework:** Gin Gonic (`github.com/gin-gonic/gin`)
*   **Orchestration:** Docker (Multi-stage build: Go builder -> Alpine runner) & Docker Compose.
*   **Persistence:** In-Memory (`map[string]*Game` protected by `sync.RWMutex`). NO external databases (Postgres/Redis).
*   **Frontend:** HTML5, CSS, Vanilla JS (Fetch API).
*   **Testing:** Go standard `testing` package.

## 3. Game Rules (Domain Logic)
**CRITICAL:** strictly adhere to these rules in `game/engine.go`.
1.  **Deck:** Standard 52-card deck. Shuffled at the start of **every** game.
2.  **Values:**
    *   2-9: Face Value
    *   10, J, Q, K: Value of 10
    *   Ace: 1 or 11 (dynamic, whichever prevents busting).
3.  **Dealer Rules (The Bot):**
    *   **MUST HIT** on Soft 17 (Ace + 6 treated as 17) or any total < 17.
    *   **MUST STAND** on Hard 17 or higher.
4.  **Win Conditions:** Higher score ≤ 21 wins. > 21 is a Bust.

## 4. Architecture & Directory Structure
The project **must** follow this exact structure:

```text
blackjack-api/
├── Dockerfile              # Multi-stage: Build in golang:alpine -> Run in alpine
├── docker-compose.yml      # Service definition (Port 8080)
├── go.mod                  # Module definition
├── main.go                 # Entry point & Router
├── game/                   # DOMAIN LOGIC
│   ├── deck.go             # Deck generation & shuffling
│   ├── engine.go           # Scoring logic & Soft 17 rules
│   ├── models.go           # Structs: Card, Hand, GameState
│   └── store.go            # Thread-safe storage (Mutex + Map)
├── handlers/               # HTTP LAYER
│   └── controller.go       # Handlers (Start, Hit, Stand)
└── web/                    # FRONTEND
    ├── index.html          # UI
    └── app.js              # Client logic
```

## 5. Development Roadmap

### Phase 1: Foundation & Infrastructure
- [ ] **1.1 Project Init:** `go mod init`, Create directory structure.
- [ ] **1.2 Docker Setup:** Create `Dockerfile` (multi-stage) and `docker-compose.yml`.

### Phase 2: Domain Logic (TDD Approach)
- [ ] **2.1 Models:** Define `Card`, `Hand`, `GameState` in `game/models.go`.
- [ ] **2.2 Deck Logic:** Implement `NewDeck` and `Shuffle` in `game/deck.go`. Write unit tests.
- [ ] **2.3 Game Engine:** Implement scoring, Ace handling, and Dealer rules in `game/engine.go`. Write unit tests.
- [ ] **2.4 In-Memory Store:** Implement thread-safe storage in `game/store.go`.

### Phase 3: API Implementation
- [ ] **3.1 Handler Skeleton:** Create `StartGame` and `PerformAction` stubs in `handlers/controller.go`. Define DTOs (`GameResponse`).
- [ ] **3.2 Start Game Logic:** Implement `POST /api/games`. Deal initial cards (hide dealer's 2nd card).
- [ ] **3.3 Action Logic (Hit):** Implement `POST /api/games/:id/action` (Hit). Handle busts.
- [ ] **3.4 Action Logic (Stand):** Implement `POST /api/games/:id/action` (Stand). Trigger Dealer AI loop and determine winner.

### Phase 4: Frontend
- [ ] **4.1 UI Structure:** `web/index.html` with game board, hand containers, and controls.
- [ ] **4.2 Client Logic:** `web/app.js` to handle API communication (Start, Hit, Stand) and DOM updates.

### Phase 5: Verification & Delivery
- [ ] **5.1 Integration Check:** Verify full game flow (Start -> Hit/Stand -> Win/Loss).
- [ ] **5.2 Final Polish:** Ensure code style (gofmt), comments, and README.

### Phase 6: Advanced UI/UX (Casino Polish)
- [ ] **6.1 Visual Overhaul:** Update CSS for a "Casino Table" aesthetic (Green felt background).
- [ ] **6.2 Realistic Cards:** Implement card styling with shadows (`box-shadow`), rounded corners, and suit icons/images to look like real physical cards.
- [ ] **6.3 Animations:** Add CSS keyframe animations for:
    -   **Shuffling:** Visual deck shuffle effect.
    -   **Dealing:** Cards sliding into position.
- [ ] **6.4 Game States:** Create distinct, visually appealing overlay screens for "You Won!" and "Game Over".

### Phase 7: Betting System (Economy)
- [ ] **7.1 Backend Models:** Update `GameState` to track `PlayerTokens` (Session based, start = 100) and `CurrentBet`.
- [ ] **7.2 API Updates:**
    -   Modify `POST /api/games` to accept `betAmount` (Max 10).
    -   Validate funds (Balance >= Bet). Deduct bet on start.
- [ ] **7.3 Payout Logic:** Update `Stand` logic to calculate winnings:
    -   Win: Return Bet * 2.
    -   Blackjack: Return Bet * 2.5 (3:2 payout).
    -   Push: Return Bet.
- [ ] **7.4 Frontend Integration:** Add UI for "Place Bet" (Input/Slider) and "Token Balance" display. Block play if balance is 0.

## 6. AI Agent Protocols
*   **Atomic Changes:** Implement one phase or sub-phase at a time.
*   **Testing:** Every new or modified code block MUST be covered by unit tests. Run `go test ./...` regularly and ensure all existing tests pass after any change.
*   **Consistency:** Maintain strict separation of concerns (Handlers layer vs Domain layer).
*   **Safety:** Explain critical file system changes before execution.
