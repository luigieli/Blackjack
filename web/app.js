let gameId = null;
const API_URL = '/api/games';

// Initialize Player ID
let playerId = localStorage.getItem('bj_player_id');
if (!playerId) {
    // Simple UUID generator
    playerId = 'xxxxxxxx-xxxx-4xxx-yxxx-xxxxxxxxxxxx'.replace(/[xy]/g, function(c) {
        var r = Math.random() * 16 | 0, v = c == 'x' ? r : (r & 0x3 | 0x8);
        return v.toString(16);
    });
    localStorage.setItem('bj_player_id', playerId);
}

// Check if there was a previous game state or balance in localStorage?
// No, we rely on the backend. But we should fetch initial balance perhaps?
// For now, balance will update when we start a game or (todo) separate endpoint.
// We'll just wait for the first game.

async function startGame() {
    const betInput = document.getElementById('bet-amount');
    const betAmount = parseInt(betInput.value, 10);

    if (isNaN(betAmount) || betAmount < 1 || betAmount > 10) {
        alert("Please enter a bet between 1 and 10.");
        return;
    }

    try {
        const response = await fetch(API_URL, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Player-ID': playerId
            },
            body: JSON.stringify({ bet_amount: betAmount })
        });

        if (!response.ok) {
            const err = await response.json();
            throw new Error(err.error || 'Failed to start game');
        }

        const data = await response.json();

        gameId = data.id;

        // Hide Betting Controls, Show Game Area
        document.getElementById('betting-controls').classList.add('hidden'); // We might want to hide just the button or disable inputs
        document.getElementById('game-area').classList.remove('hidden');
        document.getElementById('restart-btn').classList.add('hidden');
        document.getElementById('top-bar').classList.remove('hidden'); // Ensure top bar is visible

        // Enable controls
        enableControls(true);

        // Reset containers (IMPORTANT for New Game animation)
        document.getElementById('dealer-cards').innerHTML = '';
        document.getElementById('player-cards').innerHTML = '';
        document.getElementById('current-bet-display').classList.remove('hidden');

        updateUI(data);
    } catch (error) {
        console.error(error);
        alert(error.message);
    }
}

async function hit() {
    if (!gameId) return;
    try {
        const response = await fetch(`${API_URL}/${gameId}/action`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Player-ID': playerId // Good practice to send it, though not strictly needed for action logic if store has it
            },
            body: JSON.stringify({ action: 'hit' })
        });
        const data = await response.json();
        updateUI(data);
    } catch (error) {
        console.error(error);
    }
}

async function stand() {
    if (!gameId) return;
    try {
        const response = await fetch(`${API_URL}/${gameId}/action`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Player-ID': playerId
            },
            body: JSON.stringify({ action: 'stand' })
        });
        const data = await response.json();
        updateUI(data);
    } catch (error) {
        console.error(error);
    }
}

function resetGame() {
    // Return to betting screen
    document.getElementById('game-area').classList.add('hidden');
    document.getElementById('betting-controls').classList.remove('hidden');
    document.getElementById('current-bet-display').classList.add('hidden');
    document.getElementById('status').innerText = 'Place your bet';
    document.getElementById('status').style.color = 'white';
}

function updateUI(gameState) {
    const dealerContainer = document.getElementById('dealer-cards');
    const playerContainer = document.getElementById('player-cards');
    const statusDiv = document.getElementById('status');
    const dealerScoreSpan = document.getElementById('dealer-score');
    const playerScoreSpan = document.getElementById('player-score');
    const balanceSpan = document.getElementById('player-balance');
    const currentBetSpan = document.getElementById('current-bet');

    // Update Balance & Bet
    if (gameState.player_balance !== undefined) {
        balanceSpan.innerText = gameState.player_balance;
    }
    if (gameState.current_bet !== undefined) {
        currentBetSpan.innerText = gameState.current_bet;
    }

    // Update Hands with Animation Logic
    updateHand(playerContainer, gameState.player_hand.cards);
    updateHand(dealerContainer, gameState.dealer_hand.cards);

    playerScoreSpan.innerText = gameState.player_hand.score;

    // Determine Dealer Score Display
    let dealerScoreDisplay = gameState.dealer_hand.score;
    // Check if any card is hidden (rank is empty)
    const hasHidden = gameState.dealer_hand.cards.some(c => c.rank === "");
    if (hasHidden) {
        dealerScoreDisplay = "?";
    }
    dealerScoreSpan.innerText = dealerScoreDisplay;

    // Update Status Message
    statusDiv.innerText = formatStatus(gameState.status);

    // Handle Game Over
    if (gameState.status !== 'PlayerTurn') {
        enableControls(false);
        document.getElementById('restart-btn').classList.remove('hidden');

        // Highlight status
        if (gameState.status === 'PlayerWon' || gameState.status === 'DealerBust') {
            statusDiv.style.color = '#5cb85c'; // Green
        } else if (gameState.status === 'DealerWon' || gameState.status === 'PlayerBust') {
            statusDiv.style.color = '#d9534f'; // Red
        } else {
            statusDiv.style.color = '#f0ad4e'; // Orange (Push)
        }
    } else {
         statusDiv.style.color = 'white';
    }
}

function updateHand(container, cards) {
    const currentElements = container.querySelectorAll('.card-wrapper');
    const currentCount = currentElements.length;

    // 1. Add new cards
    if (cards.length > currentCount) {
        for (let i = currentCount; i < cards.length; i++) {
            const cardData = cards[i];
            const el = createCardElement(cardData);

            // Add Fly-In Animation
            el.classList.add('fly-in');

            // Stagger animations slightly if multiple cards added at once (start game)
            if (cards.length - currentCount > 1) {
                el.style.animationDelay = `${(i - currentCount) * 0.1}s`;
            }

            container.appendChild(el);
        }
    }

    // 2. Update existing cards (Check for Flip)
    for (let i = 0; i < currentCount; i++) {
        const cardData = cards[i];
        const wrapper = currentElements[i];
        const inner = wrapper.querySelector('.card-inner');
        const frontFace = inner.querySelector('.card-face.front');

        // Check if previously hidden (flipped) and now revealed
        const isCurrentlyFlipped = inner.classList.contains('flipped');
        const shouldBeHidden = (cardData.rank === "");

        if (isCurrentlyFlipped && !shouldBeHidden) {
            // Reveal!
            // Update the front face content
            updateCardFaceContent(frontFace, cardData);

            // Remove the flipped class to rotate back to 0deg
            inner.classList.remove('flipped');
        } else if (!isCurrentlyFlipped && shouldBeHidden) {
            // Hide (Rare case, maybe reset?)
            inner.classList.add('flipped');
        }

        // Ensure content matches if not hidden
        if (!shouldBeHidden) {
             updateCardFaceContent(frontFace, cardData);
        }
    }
}

function createCardElement(card) {
    // Structure:
    // .card-wrapper (fly-in)
    //   .card-inner (transform-style)
    //     .card-face.front
    //     .card-face.back

    const wrapper = document.createElement('div');
    wrapper.className = 'card-wrapper';

    const inner = document.createElement('div');
    inner.className = 'card-inner';

    // If card rank is empty, it's hidden -> Start flipped
    if (card.rank === "") {
        inner.classList.add('flipped');
    }

    // Front Face
    const front = document.createElement('div');
    updateCardFaceContent(front, card); // helper to set classes/content

    // Back Face
    const back = document.createElement('div');
    back.className = 'card-face back';

    inner.appendChild(front);
    inner.appendChild(back);
    wrapper.appendChild(inner);

    return wrapper;
}

function updateCardFaceContent(el, card) {
    el.className = 'card-face front';
    if (card.rank === "") return; // Empty content for hidden card initially

    const suitLower = card.suit ? card.suit.toLowerCase() : '';
    el.classList.add(suitLower);

    el.setAttribute('data-rank', getShortRank(card.rank));
    el.setAttribute('data-suit', getSuitSymbol(card.suit));
    el.innerText = getSuitSymbol(card.suit);
}

function getShortRank(rank) {
    if (!rank) return '';
    if (rank === '10') return '10';
    return rank.charAt(0);
}

function getSuitSymbol(suit) {
    switch (suit) {
        case 'Hearts': return '♥';
        case 'Diamonds': return '♦';
        case 'Clubs': return '♣';
        case 'Spades': return '♠';
        default: return '';
    }
}

function formatStatus(status) {
    switch (status) {
        case 'PlayerTurn': return 'Your Turn';
        case 'DealerTurn': return 'Dealer\'s Turn';
        case 'PlayerWon': return 'You Win!';
        case 'DealerWon': return 'Dealer Wins!';
        case 'Push': return 'Push (Tie)';
        default: return status;
    }
}

function enableControls(enabled) {
    document.getElementById('hit-btn').disabled = !enabled;
    document.getElementById('stand-btn').disabled = !enabled;
}
