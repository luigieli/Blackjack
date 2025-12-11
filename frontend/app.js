let gameId = null;
let playerId = localStorage.getItem('blackjack_player_id');

const API_BASE = '/api';

// Initialize Session on Load
window.addEventListener('load', async () => {
    if (!playerId) {
        await createPlayer();
    } else {
        await refreshPlayer();
    }
});

async function createPlayer() {
    try {
        const res = await fetch(`${API_BASE}/players`, { method: 'POST' });
        const data = await res.json();
        playerId = data.id;
        localStorage.setItem('blackjack_player_id', playerId);
        updateBalanceDisplay(data.balance);
    } catch (e) {
        console.error("Failed to create player", e);
    }
}

async function refreshPlayer() {
    try {
        const res = await fetch(`${API_BASE}/players/${playerId}`);
        if (res.status === 404) {
            // ID invalid, recreate
            localStorage.removeItem('blackjack_player_id');
            await createPlayer();
            return;
        }
        const data = await res.json();
        updateBalanceDisplay(data.balance);
    } catch (e) {
        console.error("Failed to refresh player", e);
    }
}

async function resetBank() {
    try {
        const res = await fetch(`${API_BASE}/players/${playerId}/reset`, { method: 'POST' });
        const data = await res.json();
        updateBalanceDisplay(data.balance);
        document.getElementById('reset-btn').style.display = 'none';
        document.getElementById('start-btn').disabled = false;
        document.getElementById('message').innerText = "";
    } catch (e) {
        console.error(e);
    }
}

async function startGame() {
    if (!playerId) await createPlayer();

    const betInput = document.getElementById('bet-amount');
    const betAmount = parseInt(betInput.value, 10);

    if (isNaN(betAmount) || betAmount < 1 || betAmount > 10) {
        showMessage("Bet must be between 1 and 10");
        return;
    }

    try {
        const response = await fetch(`${API_BASE}/games`, {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify({
                player_id: playerId,
                bet_amount: betAmount
            })
        });

        if (!response.ok) {
            const err = await response.json();
            throw new Error(err.error || 'Failed to start game');
        }

        const data = await response.json();
        gameId = data.id;

        // UI Transition
        document.getElementById('setup-area').style.display = 'none';
        document.getElementById('game-area').style.display = 'block';
        document.getElementById('next-hand-btn').style.display = 'none';
        document.getElementById('payout-msg').innerText = '';

        updateUI(data);
        enableControls(true);
    } catch (error) {
        console.error(error);
        showMessage(error.message);
        // Check if balance 0
        refreshPlayer();
    }
}

async function hit() {
    if (!gameId) return;
    try {
        const response = await fetch(`${API_BASE}/games/${gameId}/action`, {
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
        const response = await fetch(`${API_BASE}/games/${gameId}/action`, {
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
    const payoutMsg = document.getElementById('payout-msg');

    // Update Balance
    if (gameState.player_balance !== undefined) {
        updateBalanceDisplay(gameState.player_balance);
    }

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
        document.getElementById('next-hand-btn').style.display = 'inline-block';

        if (gameState.payout > 0) {
            payoutMsg.innerText = `You won $${gameState.payout}!`;
        } else if (gameState.status === 'Push') {
            payoutMsg.innerText = `Push. Bet returned.`;
        } else {
            payoutMsg.innerText = `House Wins.`;
        }
    }
}

function resetUIForNewGame() {
    document.getElementById('game-area').style.display = 'none';
    document.getElementById('setup-area').style.display = 'block';
    showMessage('');
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

function updateBalanceDisplay(balance) {
    document.getElementById('balance').innerText = `$${balance}`;
    if (balance === 0) {
        document.getElementById('reset-btn').style.display = 'inline-block';
        document.getElementById('start-btn').disabled = true;
        showMessage("You are out of money!");
    } else {
        document.getElementById('reset-btn').style.display = 'none';
        document.getElementById('start-btn').disabled = false;
    }
}

function showMessage(msg) {
    document.getElementById('message').innerText = msg;
}
