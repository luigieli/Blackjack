let gameId = null;

const API_URL = '/api/games';

async function startGame() {
    try {
        const response = await fetch(API_URL, { method: 'POST' });
        if (!response.ok) throw new Error('Failed to start game');
        const data = await response.json();

        gameId = data.id;
        document.getElementById('start-btn').style.display = 'none';
        document.getElementById('game-area').style.display = 'block';
        document.getElementById('restart-btn').style.display = 'none';

        updateUI(data);
        enableControls(true);
    } catch (error) {
        console.error(error);
        alert('Error starting game');
    }
}

async function hit() {
    if (!gameId) return;
    try {
        const response = await fetch(`${API_URL}/${gameId}/action`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
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
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({ action: 'stand' })
        });
        const data = await response.json();
        updateUI(data);
    } catch (error) {
        console.error(error);
    }
}

function updateUI(gameState) {
    const dealerContainer = document.getElementById('dealer-cards');
    const playerContainer = document.getElementById('player-cards');
    const statusDiv = document.getElementById('status');
    const dealerScoreSpan = document.getElementById('dealer-score');
    const playerScoreSpan = document.getElementById('player-score');

    // Update Player
    playerContainer.innerHTML = '';
    gameState.player_hand.cards.forEach(card => {
        playerContainer.appendChild(createCardElement(card));
    });
    playerScoreSpan.innerText = gameState.player_hand.score;

    // Update Dealer
    dealerContainer.innerHTML = '';
    let dealerScoreDisplay = gameState.dealer_hand.score;
    gameState.dealer_hand.cards.forEach(card => {
        if (card.rank === "") { // Masked card
            const el = document.createElement('div');
            el.className = 'card hidden-card';
            el.innerText = '?';
            dealerContainer.appendChild(el);
            dealerScoreDisplay = "?";
        } else {
            dealerContainer.appendChild(createCardElement(card));
        }
    });
    dealerScoreSpan.innerText = dealerScoreDisplay;

    // Status
    statusDiv.innerText = `Status: ${gameState.status}`;

    if (gameState.status !== 'PlayerTurn') {
        enableControls(false);
        document.getElementById('restart-btn').style.display = 'inline-block';
    }
}

function createCardElement(card) {
    const el = document.createElement('div');
    el.className = 'card';
    el.innerText = `${card.rank} ${getSuitSymbol(card.suit)}`;
    return el;
}

function getSuitSymbol(suit) {
    switch (suit) {
        case 'Hearts': return '♥';
        case 'Diamonds': return '♦';
        case 'Clubs': return '♣';
        case 'Spades': return '♠';
        default: return suit;
    }
}

function enableControls(enabled) {
    document.getElementById('hit-btn').disabled = !enabled;
    document.getElementById('stand-btn').disabled = !enabled;
}
