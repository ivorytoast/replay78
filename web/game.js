let ws;
let gameState = {
    board: Array(3).fill(null).map(() => Array(3).fill({ player: 0, power: 0 })),
    currentPlayer: 1,
    done: false,
    currentPhase: 0,        // 0=Assignment, 1=Movement
    movementTaken: false
};
let selectedCell = null;

// Helper function to check if two cells are adjacent (orthogonal only)
function isAdjacent(fromRow, fromCol, toRow, toCol) {
    const rowDiff = Math.abs(fromRow - toRow);
    const colDiff = Math.abs(fromCol - toCol);
    return rowDiff + colDiff === 1;
}

// Get all adjacent cells for a given position
function getAdjacentCells(row, col) {
    const adjacent = [];
    // Up
    if (row > 0) adjacent.push({row: row - 1, col: col});
    // Down
    if (row < 2) adjacent.push({row: row + 1, col: col});
    // Left
    if (col > 0) adjacent.push({row: row, col: col - 1});
    // Right
    if (col < 2) adjacent.push({row: row, col: col + 1});
    return adjacent;
}

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
                player2Lines: data.player2Lines,
                currentPhase: data.currentPhase || 0,
                movementTaken: data.movementTaken || false
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

            // Highlight valid move targets when a piece is selected
            if (selectedCell) {
                const selectedPlayer = gameState.board[selectedCell.row][selectedCell.col].player;
                const selectedPower = gameState.board[selectedCell.row][selectedCell.col].power;

                // Only highlight for different-cell operations (moves/attacks)
                if (selectedCell.row !== i || selectedCell.col !== j) {
                    if (selectedPlayer === gameState.currentPlayer && selectedPower > 0) {
                        // Check if this cell is adjacent
                        if (isAdjacent(selectedCell.row, selectedCell.col, i, j)) {
                            cell.classList.add('valid-target');
                        }
                    }
                }
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

    const currentPhase = gameState.currentPhase;
    const isAssignmentPhase = currentPhase === 0;
    const isMovementPhase = currentPhase === 1;

    if (!selectedCell) {
        // First click - select source
        selectedCell = { row, col };
        const player = gameState.board[row][col].player;
        const currentPlayerName = gameState.currentPlayer === 1 ? 'X' : 'O';

        if (player === 0) {
            if (isAssignmentPhase) {
                setStatus(`Selected empty cell (${row},${col}) - Click again to place ${currentPlayerName}`);
            } else {
                setStatus(`Cannot place pieces during movement phase - select your piece to move`);
                selectedCell = null;
                renderBoard();
                return;
            }
        } else if (player === gameState.currentPlayer) {
            if (isAssignmentPhase) {
                setStatus(`Selected your piece at (${row},${col}) - Click again to power up`);
            } else {
                setStatus(`Selected your piece at (${row},${col}) - Click where to move/attack/combine`);
            }
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

        // Phase-specific validation
        const isSameCell = (fromRow === toRow && fromCol === toCol);

        if (isAssignmentPhase) {
            // In assignment phase: only allow same-cell operations
            if (!isSameCell) {
                setStatus('During assignment phase, you can only place or power up pieces (click same cell)');
                selectedCell = null;
                renderBoard();
                return;
            }
        } else if (isMovementPhase) {
            // In movement phase: only allow different-cell operations
            if (isSameCell) {
                setStatus('During movement phase, you cannot place or power up (click different cell)');
                selectedCell = null;
                renderBoard();
                return;
            }

            // Check if already made movement action
            if (gameState.movementTaken) {
                setStatus('You have already made your movement action this turn - click "End Turn"');
                selectedCell = null;
                renderBoard();
                return;
            }

            // Validate adjacency for movement phase
            const selectedPlayer = gameState.board[fromRow][fromCol].player;
            const selectedPower = gameState.board[fromRow][fromCol].power;

            if (selectedPlayer === gameState.currentPlayer && selectedPower > 0) {
                if (!isAdjacent(fromRow, fromCol, toRow, toCol)) {
                    setStatus('Invalid move: Can only move to adjacent cells (up/down/left/right)');
                    selectedCell = null;
                    renderBoard();
                    return;
                }
            }
        }

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

function endTurn() {
    if (gameState.currentPhase !== 1) {
        setStatus('Cannot end turn - must complete assignment phase first');
        return;
    }

    ws.send(JSON.stringify({ type: 'endturn' }));
    selectedCell = null;
    setStatus('Ending turn...');
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
        // Display current phase
        const phaseText = gameState.currentPhase === 0 ? 'Assignment' : 'Movement';
        gameStatusEl.textContent = `Phase: ${phaseText}`;
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

    // Update end turn button visibility
    const endTurnBtn = document.getElementById('endTurnBtn');
    if (endTurnBtn) {
        if (gameState.currentPhase === 1 && !gameState.movementTaken && !gameState.done) {
            endTurnBtn.style.display = 'inline-block';
        } else {
            endTurnBtn.style.display = 'none';
        }
    }
}

function setStatus(message) {
    const statusEl = document.getElementById('status');

    // Add phase context to status messages
    if (!gameState.done) {
        const phasePrefix = gameState.currentPhase === 0
            ? '[Assignment Phase] '
            : '[Movement Phase] ';
        statusEl.textContent = phasePrefix + message;
    } else {
        statusEl.textContent = message;
    }
    statusEl.className = 'status';
    if (gameState.done) {
        statusEl.classList.add('game-over');
    }
}

// Initialize
connect();
