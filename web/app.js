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

async function startGame() {
    const betInput = document.getElementById('bet-amount');
    const betAmount = parseInt(betInput.value, 10);

    if (isNaN(betAmount) || betAmount < 1) {
        alert("Please enter a bet of at least 1.");
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
        document.getElementById('betting-controls').classList.add('hidden');
        document.getElementById('game-area').classList.remove('hidden');
        document.getElementById('restart-btn').classList.add('hidden');
        document.getElementById('top-bar').classList.remove('hidden'); // Ensure top bar is visible

        // Enable controls
        enableControls(true);

        // Reset containers (IMPORTANT for New Game animation)
        document.getElementById('dealer-cards').innerHTML = '';
        document.getElementById('player-cards').innerHTML = '';
        document.getElementById('split-cards').innerHTML = '';
        document.getElementById('current-bet-display').classList.remove('hidden');
        document.getElementById('split-hand-container').classList.add('hidden');
        document.getElementById('player-hand-container').classList.remove('active-hand');
        document.getElementById('split-hand-container').classList.remove('active-hand');

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
                'X-Player-ID': playerId
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

async function split() {
    if (!gameId) return;
    try {
        const response = await fetch(`${API_URL}/${gameId}/action`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'X-Player-ID': playerId
            },
            body: JSON.stringify({ action: 'split' })
        });
        const data = await response.json();

        // Clear containers to force re-render or handle carefully?
        // Actually, updateHand handles additions. But split moves a card.
        // updateHand logic might be tricky if a card is removed (moved).
        // Standard updateHand only appends or updates.
        // We should clear the containers if we detect a split happened (or just clear always?)
        // Clearing breaks animations slightly but ensures correctness.
        // Let's clear player container if split_hand is present for the first time?
        // Or simply: if split action success, clear player container to redraw properly.
        document.getElementById('player-cards').innerHTML = '';
        document.getElementById('split-cards').innerHTML = '';

        updateUI(data);
    } catch (error) {
        console.error(error);
        alert(error.message || "Failed to split");
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
    const splitContainer = document.getElementById('split-cards');

    const statusDiv = document.getElementById('status');
    const dealerScoreSpan = document.getElementById('dealer-score');
    const playerScoreSpan = document.getElementById('player-score');
    const splitScoreSpan = document.getElementById('split-score');

    const balanceSpan = document.getElementById('player-balance');
    const currentBetSpan = document.getElementById('current-bet');
    const splitBtn = document.getElementById('split-btn');

    // Update Balance & Bet
    if (gameState.player_balance !== undefined) {
        balanceSpan.innerText = gameState.player_balance;
    }
    if (gameState.current_bet !== undefined) {
        currentBetSpan.innerText = gameState.current_bet;
    }

    // Update Hands
    updateHand(playerContainer, gameState.player_hand.cards);
    updateHand(dealerContainer, gameState.dealer_hand.cards);
    playerScoreSpan.innerText = gameState.player_hand.score;

    // Split Logic
    if (gameState.split_hand) {
        document.getElementById('split-hand-container').classList.remove('hidden');
        updateHand(splitContainer, gameState.split_hand.cards);
        splitScoreSpan.innerText = gameState.split_hand.score;

        // Active Hand Indicator
        document.getElementById('player-hand-container').classList.remove('active-hand');
        document.getElementById('split-hand-container').classList.remove('active-hand');

        if (gameState.status === 'PlayerTurn') {
            if (gameState.current_hand_index === 0) {
                document.getElementById('player-hand-container').classList.add('active-hand');
            } else {
                document.getElementById('split-hand-container').classList.add('active-hand');
            }
        }
    } else {
        document.getElementById('split-hand-container').classList.add('hidden');
    }

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

    // Controls Logic
    if (gameState.status !== 'PlayerTurn') {
        enableControls(false);
        splitBtn.classList.add('hidden');
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
        // Split Button Visibility
        // Show if: PlayerTurn, no split hand yet, 2 cards in player hand, ranks match, enough balance.
        // We need balance locally? We have it in gameState.player_balance but we need to know if we can afford.
        // Also need to know if this is the FIRST action (2 cards).
        // We can just rely on the fact that if we have 2 cards and not split, we can *potentially* split.
        // We'll let the backend validate strictly, but UI should be smart.

        const canSplit = !gameState.split_hand &&
                         gameState.player_hand.cards.length === 2 &&
                         gameState.player_hand.cards[0].rank === gameState.player_hand.cards[1].rank &&
                         gameState.player_balance >= gameState.current_bet;

        if (canSplit) {
            splitBtn.classList.remove('hidden');
        } else {
            splitBtn.classList.add('hidden');
        }
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

            // Stagger animations
            if (cards.length - currentCount > 1) {
                el.style.animationDelay = `${(i - currentCount) * 0.1}s`;
            }

            container.appendChild(el);
        }
    }

    // 2. Update existing cards
    for (let i = 0; i < currentCount; i++) {
        const cardData = cards[i];
        const wrapper = currentElements[i];
        const inner = wrapper.querySelector('.card-inner');
        const frontFace = inner.querySelector('.card-face.front');

        const isCurrentlyFlipped = inner.classList.contains('flipped');
        const shouldBeHidden = (cardData.rank === "");

        if (isCurrentlyFlipped && !shouldBeHidden) {
            updateCardFaceContent(frontFace, cardData);
            inner.classList.remove('flipped');
        } else if (!isCurrentlyFlipped && shouldBeHidden) {
            inner.classList.add('flipped');
        }

        if (!shouldBeHidden) {
             updateCardFaceContent(frontFace, cardData);
        }
    }
}

function createCardElement(card) {
    const wrapper = document.createElement('div');
    wrapper.className = 'card-wrapper';

    const inner = document.createElement('div');
    inner.className = 'card-inner';

    if (card.rank === "") {
        inner.classList.add('flipped');
    }

    const front = document.createElement('div');
    updateCardFaceContent(front, card);

    const back = document.createElement('div');
    back.className = 'card-face back';

    inner.appendChild(front);
    inner.appendChild(back);
    wrapper.appendChild(inner);

    return wrapper;
}

function updateCardFaceContent(el, card) {
    el.className = 'card-face front';
    if (card.rank === "") return;

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
    document.getElementById('split-btn').disabled = !enabled;
}
