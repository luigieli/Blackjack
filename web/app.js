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

        // Reset containers (IMPORTANT for New Game animation)
        document.getElementById('dealer-cards').innerHTML = '';
        document.getElementById('player-cards').innerHTML = '';

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
