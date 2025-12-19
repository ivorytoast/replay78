let ws;
let gameState = {
    board: Array(3).fill(null).map(() => Array(3).fill({ player: 0, power: 0 })),
    currentPlayer: 1,
    done: false
};
let selectedCell = null;

// Connect to WebSocket
function connect() {
    ws = new WebSocket('ws://localhost:8080/ws');

    ws.onopen = () => {
        console.log('Connected to server');
        updateConnectionStatus(true);
        setStatus('Connected! Click a cell to place your first piece');
    };

    ws.onmessage = (event) => {
        const data = JSON.parse(event.data);
        if (data.type === 'board_state') {
            gameState = {
                board: data.board,
                currentPlayer: data.currentPlayer,
                done: data.done,
                player1PowerBank: data.player1PowerBank,
                player2PowerBank: data.player2PowerBank,
                player1Lines: data.player1Lines,
                player2Lines: data.player2Lines
            };
            renderBoard();
            updateGameInfo();
        }
    };

    ws.onclose = () => {
        console.log('Disconnected from server');
        updateConnectionStatus(false);
        setStatus('Disconnected. Refresh to reconnect.');
        setTimeout(connect, 3000); // Try to reconnect
    };

    ws.onerror = (error) => {
        console.error('WebSocket error:', error);
    };
}

function updateConnectionStatus(connected) {
    const status = document.getElementById('connectionStatus');
    status.textContent = connected ? '● Connected' : '● Disconnected';
    status.className = 'connection-status ' + (connected ? 'connected' : 'disconnected');
}

function renderBoard() {
    const boardEl = document.getElementById('board');
    boardEl.innerHTML = '';

    for (let i = 0; i < 3; i++) {
        for (let j = 0; j < 3; j++) {
            const cell = document.createElement('div');
            cell.className = 'cell';
            cell.dataset.row = i;
            cell.dataset.col = j;
            cell.dataset.player = gameState.board[i][j].player;

            const player = gameState.board[i][j].player;
            const power = gameState.board[i][j].power;

            if (player === 0) {
                cell.classList.add('empty');
                cell.innerHTML = '<span class="cell-player">·</span>';
            } else {
                const symbol = player === 1 ? 'X' : 'O';
                const powerBubble = `<div class="cell-power">${power}</div>`;
                cell.innerHTML = `${powerBubble}<span class="cell-player">${symbol}</span>`;
            }

            // Check if this cell is selected
            if (selectedCell && selectedCell.row === i && selectedCell.col === j) {
                cell.classList.add('selected');
            }

            cell.onclick = () => handleCellClick(i, j);
            boardEl.appendChild(cell);
        }
    }
}

function handleCellClick(row, col) {
    if (gameState.done) {
        setStatus('Game is over! Start a new game.');
        return;
    }

    if (!selectedCell) {
        // First click - select source
        selectedCell = { row, col };
        const player = gameState.board[row][col].player;
        const currentPlayerName = gameState.currentPlayer === 1 ? 'X' : 'O';

        if (player === 0) {
            setStatus(`Selected empty cell (${row},${col}) - Click again to place ${currentPlayerName} or click another cell`);
        } else if (player === gameState.currentPlayer) {
            setStatus(`Selected your piece at (${row},${col}) - Click where to move/attack/power-up`);
        } else {
            setStatus(`That's opponent's piece! Click your own piece (${currentPlayerName})`);
            selectedCell = null;
            renderBoard();
            return;
        }
        renderBoard();
    } else {
        // Second click - execute move
        const fromRow = selectedCell.row;
        const fromCol = selectedCell.col;
        const toRow = row;
        const toCol = col;

        sendMove(fromRow, fromCol, toRow, toCol);
        selectedCell = null;
        renderBoard();
    }
}

function sendMove(fromRow, fromCol, toRow, toCol) {
    const payload = `${fromRow} ${fromCol} ${toRow} ${toCol}`;
    ws.send(JSON.stringify({
        type: 'move',
        payload: payload
    }));
    setStatus('Processing move...');
}

function newGame() {
    ws.send(JSON.stringify({ type: 'new' }));
    selectedCell = null;
    setStatus('New game started!');
}

function refreshBoard() {
    ws.send(JSON.stringify({ type: 'show' }));
}

function updateGameInfo() {
    const currentPlayerEl = document.getElementById('currentPlayer');
    const playerName = gameState.currentPlayer === 1 ? 'X' : 'O';
    const playerClass = gameState.currentPlayer === 1 ? 'player-x' : 'player-o';
    currentPlayerEl.innerHTML = `Current: <span class="${playerClass}">${playerName}</span>`;

    const gameStatusEl = document.getElementById('gameStatus');
    if (gameState.done) {
        gameStatusEl.textContent = 'Game Over!';
    } else {
        gameStatusEl.textContent = 'In Progress';
    }

    // Update power banks display
    const player1PowerBankEl = document.getElementById('player1PowerBank');
    const player2PowerBankEl = document.getElementById('player2PowerBank');
    const player1LineBonusEl = document.getElementById('player1LineBonus');
    const player2LineBonusEl = document.getElementById('player2LineBonus');

    if (player1PowerBankEl) {
        player1PowerBankEl.textContent = gameState.player1PowerBank !== undefined ? gameState.player1PowerBank : 1;
    }
    if (player2PowerBankEl) {
        player2PowerBankEl.textContent = gameState.player2PowerBank !== undefined ? gameState.player2PowerBank : 1;
    }

    // Update line bonus badges
    if (player1LineBonusEl) {
        const p1Lines = gameState.player1Lines || 0;
        if (p1Lines > 0) {
            player1LineBonusEl.textContent = `+${p1Lines}`;
            player1LineBonusEl.classList.add('visible');
        } else {
            player1LineBonusEl.classList.remove('visible');
        }
    }
    if (player2LineBonusEl) {
        const p2Lines = gameState.player2Lines || 0;
        if (p2Lines > 0) {
            player2LineBonusEl.textContent = `+${p2Lines}`;
            player2LineBonusEl.classList.add('visible');
        } else {
            player2LineBonusEl.classList.remove('visible');
        }
    }
}

function setStatus(message) {
    const statusEl = document.getElementById('status');
    statusEl.textContent = message;
    statusEl.className = 'status';
    if (gameState.done) {
        statusEl.classList.add('game-over');
    }
}

// Initialize
connect();
