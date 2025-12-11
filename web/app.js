let gameId = null;
const API_URL = '/api/games';

async function startGame() {
    try {
        const response = await fetch(API_URL, { method: 'POST' });
        if (!response.ok) throw new Error('Failed to start game');
        const data = await response.json();

        gameId = data.id;

        // Show Game Area, Hide Start Button
        document.getElementById('start-btn').classList.add('hidden');
        document.getElementById('game-area').classList.remove('hidden');
        document.getElementById('restart-btn').classList.add('hidden');

        // Enable controls
        enableControls(true);

        updateUI(data);
    } catch (error) {
        console.error(error);
        alert('Error starting game. Is the backend running?');
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

    // Update Player Hand
    playerContainer.innerHTML = '';
    gameState.player_hand.cards.forEach(card => {
        playerContainer.appendChild(createCardElement(card));
    });
    playerScoreSpan.innerText = gameState.player_hand.score;

    // Update Dealer Hand
    dealerContainer.innerHTML = '';
    let dealerScoreDisplay = gameState.dealer_hand.score;
    let isMasked = false;

    gameState.dealer_hand.cards.forEach(card => {
        if (card.rank === "") { // Masked card
            const el = document.createElement('div');
            el.className = 'card hidden-card';
            dealerContainer.appendChild(el);
            isMasked = true;
        } else {
            dealerContainer.appendChild(createCardElement(card));
        }
    });

    if (isMasked) {
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

function createCardElement(card) {
    const el = document.createElement('div');
    const suitLower = card.suit.toLowerCase();
    el.className = `card ${suitLower}`;

    // Set data attributes for pseudo-elements (corners)
    el.setAttribute('data-rank', getShortRank(card.rank));
    el.setAttribute('data-suit', getSuitSymbol(card.suit));

    // Main center content
    el.innerText = getSuitSymbol(card.suit);

    return el;
}

function getShortRank(rank) {
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
