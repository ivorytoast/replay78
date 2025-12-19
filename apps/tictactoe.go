package apps

import (
	"fmt"
	"replay78/assert"
	"replay78/engine"
	"replay78/states"
)

type TicTacToeApp struct {
	engine *engine.Engine
}

func NewTicTacToeApp(e *engine.Engine) *TicTacToeApp {
	return &TicTacToeApp{engine: e}
}

func (t *TicTacToeApp) Topics() []string {
	return []string{"ttt"}
}

func (t *TicTacToeApp) OnEvent(event []string) {
	topic := event[0]
	action := event[1]
	payload := event[2]

	assert.Is(topic != "")
	assert.Is(action != "")
	assert.Is(t != nil)

	if topic == "ttt" {
		switch action {
		case "new":
			t.reset()
			t.engine.Out("new game command processed")
			t.showBoard()
		case "show":
			t.engine.Out("show command processed")
			t.showBoard()
		case "move":
			var fromRow, fromCol, toRow, toCol int
			_, err := fmt.Sscanf(payload, "%d %d %d %d", &fromRow, &fromCol, &toRow, &toCol)
			if err != nil {
				t.engine.Out("invalid user action")
				return
			}
			if t.engine.TTT().IsDone() {
				t.engine.Out("Move rejected - game ended")
			} else {
				moveResult := t.makeMove(fromRow, fromCol, toRow, toCol)
				if moveResult != "" {
					t.engine.Out(moveResult)
					if t.engine.TTT().IsDone() {
						winner := t.getWinner()
						if winner == 0 {
							t.engine.Out("Game Over: Tie")
						} else {
							t.engine.Out(fmt.Sprintf("Game Over: Player %d wins", winner))
						}
					}
					t.showBoard()
				} else {
					t.engine.Out("Invalid move")
				}
			}
		}
	}
}

func (t *TicTacToeApp) makeMove(fromRow, fromCol, toRow, toCol int) string {
	state := t.engine.TTT()
	currentPlayer := state.GetCurrentPlayer()

	// Check bounds
	if fromRow < 0 || fromRow > 2 || fromCol < 0 || fromCol > 2 ||
		toRow < 0 || toRow > 2 || toCol < 0 || toCol > 2 {
		return ""
	}

	assert.Is(state != nil)
	assert.Is(state.GetBoard() != nil)
	assert.Is(len(state.GetBoard()) > fromRow)
	assert.Is(len(state.GetBoard()[fromRow]) > fromCol)
	assert.Is(len(state.GetBoard()) > toRow)
	assert.Is(len(state.GetBoard()[toRow]) > toCol)

	board := state.GetBoard()
	fromCell := board[fromRow][fromCol]
	toCell := board[toRow][toCol]

	// Count lines before move
	player1LinesBefore := t.CountLines(1)
	player2LinesBefore := t.CountLines(2)

	var result string

	// Special case: placing/powering up at same location
	if fromRow == toRow && fromCol == toCol {
		if fromCell.Player == 0 {
			// Check power bank before allowing placement
			currentPowerBank := state.GetPowerBank(currentPlayer)
			if currentPowerBank <= 0 {

				return "" // Reject move
			}
			// Deduct from power bank
			state.IncrementPowerBank(currentPlayer, -1)
			// Placing initial piece
			state.SetCell(fromRow, fromCol, currentPlayer)
			result = fmt.Sprintf("Placed piece at (%d,%d) (bank: %d)", fromRow, fromCol, state.GetPowerBank(currentPlayer))
		} else if fromCell.Player == currentPlayer {
			// Check power bank before allowing power-up
			currentPowerBank := state.GetPowerBank(currentPlayer)
			if currentPowerBank <= 0 {
				return "" // Reject move
			}
			// Deduct from power bank
			state.IncrementPowerBank(currentPlayer, -1)
			state.IncrementCellPower(fromRow, fromCol)
			result = fmt.Sprintf("Power up: (%d,%d) now has power %d (bank: %d)",
				fromRow, fromCol, fromCell.Power+1, state.GetPowerBank(currentPlayer))
		} else {
			// Can't power up opponent's piece
			return ""
		}
	} else {
		// Validate source cell is owned by current player (for all non-same-cell moves)
		if fromCell.Player != currentPlayer {
			return ""
		}

		// Validate source cell has power > 0 for moves/attacks
		if fromCell.Power <= 0 {
			return ""
		}

		// Validate adjacency for moves/attacks (different-cell operations only)
		if !t.isAdjacent(fromRow, fromCol, toRow, toCol) {
			return ""
		}

		if toCell.Player == 0 {
			// Moving to empty cell
			state.SetCell(toRow, toCol, currentPlayer)
			state.SetCellPower(toRow, toCol, fromCell.Power)
			state.ClearCell(fromRow, fromCol)
			result = fmt.Sprintf("Move: (%d,%d) -> (%d,%d)", fromRow, fromCol, toRow, toCol)
		} else if toCell.Player == currentPlayer {
			// Combining two adjacent pieces of the same player
			combinedPower := fromCell.Power + toCell.Power
			state.SetCellPower(toRow, toCol, combinedPower)
			state.ClearCell(fromRow, fromCol)
			result = fmt.Sprintf("Combine: (%d,%d) + (%d,%d) -> power %d", fromRow, fromCol, toRow, toCol, combinedPower)
		} else {
			// Combat with opponent
			attackPower := fromCell.Power
			defensePower := toCell.Power

			if attackPower > defensePower {
				// Attacker wins
				newPower := attackPower - defensePower
				state.SetCell(toRow, toCol, currentPlayer)
				state.SetCellPower(toRow, toCol, newPower)
				state.ClearCell(fromRow, fromCol)
				result = fmt.Sprintf("Combat: (%d,%d) defeats (%d,%d) [%d vs %d] - Captured with power %d",
					fromRow, fromCol, toRow, toCol, attackPower, defensePower, newPower)
			} else if attackPower == defensePower {
				// Draw - both go to power 0 but keep positions
				state.SetCellPower(fromRow, fromCol, 0)
				state.SetCellPower(toRow, toCol, 0)
				result = fmt.Sprintf("Combat: (%d,%d) draws with (%d,%d) [%d vs %d] - Both reduced to power 0",
					fromRow, fromCol, toRow, toCol, attackPower, defensePower)
			} else {
				// Attacker loses
				newDefensePower := defensePower - attackPower
				state.ClearCell(fromRow, fromCol)
				state.SetCellPower(toRow, toCol, newDefensePower)
				result = fmt.Sprintf("Combat: (%d,%d) defeated by (%d,%d) [%d vs %d] - Attacker eliminated, defender at power %d",
					fromRow, fromCol, toRow, toCol, attackPower, defensePower, newDefensePower)
			}
		}
	}

	// Count lines after move (for display purposes only, no immediate bonus)
	player1LinesAfter := t.CountLines(1)
	player2LinesAfter := t.CountLines(2)

	// Display line change messages (no power bank changes here)
	if currentPlayer == 1 {
		linesGained := player1LinesAfter - player1LinesBefore
		linesLost := player2LinesBefore - player2LinesAfter
		if linesGained > 0 {
			result += fmt.Sprintf(" [Formed %d line(s)!]", linesGained)
		}
		if linesLost > 0 {
			result += fmt.Sprintf(" [Opponent lost %d line(s)!]", linesLost)
		}
	} else {
		linesGained := player2LinesAfter - player2LinesBefore
		linesLost := player1LinesBefore - player1LinesAfter
		if linesGained > 0 {
			result += fmt.Sprintf(" [Formed %d line(s)!]", linesGained)
		}
		if linesLost > 0 {
			result += fmt.Sprintf(" [Opponent lost %d line(s)!]", linesLost)
		}
	}

	// Check for game end
	if t.checkWinner() {
		state.SetDone(true)
	}

	// Mark current player's first turn as done
	state.SetFirstTurnDone(currentPlayer)

	// Switch player
	nextPlayer := 3 - currentPlayer
	state.SetCurrentPlayer(nextPlayer)

	// Only grant turn bonus to next player if BOTH players have completed their first turn
	bothPlayersHadFirstTurn := state.IsFirstTurnDone(1) && state.IsFirstTurnDone(2)
	if bothPlayersHadFirstTurn {
		// Base turn bonus: +1
		// Line bonus: +1 per line the next player has
		nextPlayerLines := t.CountLines(nextPlayer)
		turnBonus := 1 + nextPlayerLines
		state.IncrementPowerBank(nextPlayer, turnBonus)
	}

	return result
}

func (t *TicTacToeApp) isAdjacent(fromRow, fromCol, toRow, toCol int) bool {
	rowDiff := abs(fromRow - toRow)
	colDiff := abs(fromCol - toCol)
	return rowDiff+colDiff == 1
}

func abs(x int) int {
	if x < 0 {
		return -x
	}
	return x
}

func (t *TicTacToeApp) checkWinner() bool {
	player1Count := 0
	player2Count := 0

	b := t.engine.TTT().GetBoard()
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if b[i][j].Player == 1 {
				player1Count++
			} else if b[i][j].Player == 2 {
				player2Count++
			}
		}
	}

	return player1Count == 9 || player2Count == 9
}

func (t *TicTacToeApp) isFull() bool {
	b := t.engine.TTT().GetBoard()
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if b[i][j].Player == 0 {
				return false
			}
		}
	}
	return true
}

func (t *TicTacToeApp) getWinner() int {
	player1Count := 0
	player2Count := 0

	b := t.engine.TTT().GetBoard()
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			if b[i][j].Player == 1 {
				player1Count++
			} else if b[i][j].Player == 2 {
				player2Count++
			}
		}
	}

	if player1Count == 9 {
		return 1
	} else if player2Count == 9 {
		return 2
	}
	return 0
}

func (t *TicTacToeApp) reset() {
	state := t.engine.TTT()
	state.Board = [][]states.Cell{
		{{Player: 0, Power: 0}, {Player: 0, Power: 0}, {Player: 0, Power: 0}},
		{{Player: 0, Power: 0}, {Player: 0, Power: 0}, {Player: 0, Power: 0}},
		{{Player: 0, Power: 0}, {Player: 0, Power: 0}, {Player: 0, Power: 0}},
	}
	state.SetCurrentPlayer(1)
	state.SetDone(false)
	state.SetPowerBank(1, 1)
	state.SetPowerBank(2, 1)
	state.Player1FirstTurnDone = false
	state.Player2FirstTurnDone = false
}

func (t *TicTacToeApp) showBoard() {
	b := t.engine.TTT().GetBoard()
	board := ""
	boardFlattened := ""
	for i := 0; i < 3; i++ {
		for j := 0; j < 3; j++ {
			var cell string
			switch b[i][j].Player {
			case 0:
				cell = "."
			case 1:
				if b[i][j].Power == 1 {
					cell = "X"
				} else {
					cell = fmt.Sprintf("X%d", b[i][j].Power)
				}
			case 2:
				if b[i][j].Power == 1 {
					cell = "O"
				} else {
					cell = fmt.Sprintf("O%d", b[i][j].Power)
				}
			}
			board += cell
			boardFlattened += cell
			if j < 2 {
				board += " "
			}
		}
		if i < 2 {
			board += "\n"
		}
	}

	if t.engine.TTT().IsDone() {
		winner := t.getWinner()
		if winner == 0 {
			board += "\n\n*** GAME OVER - TIE ***"
			board += "\nStart a new game to continue playing."
		} else {
			playerName := "X"
			if winner == 2 {
				playerName = "O"
			}
			board += fmt.Sprintf("\n\n*** GAME OVER - PLAYER %s WINS! ***", playerName)
			board += "\nStart a new game to continue playing."
		}
	}

	t.engine.Out(boardFlattened)
	t.engine.Out(fmt.Sprintf("Player 1 (X) Power Bank: %d", t.engine.TTT().GetPowerBank(1)))
	t.engine.Out(fmt.Sprintf("Player 2 (O) Power Bank: %d", t.engine.TTT().GetPowerBank(2)))
	t.engine.Out(fmt.Sprintf("Current Turn: Player %d", t.engine.TTT().GetCurrentPlayer()))
}

func (t *TicTacToeApp) CountLines(player int) int {
	b := t.engine.TTT().GetBoard()
	count := 0

	// Check rows
	for i := 0; i < 3; i++ {
		if b[i][0].Player == player && b[i][1].Player == player && b[i][2].Player == player {
			count++
		}
	}

	// Check columns
	for i := 0; i < 3; i++ {
		if b[0][i].Player == player && b[1][i].Player == player && b[2][i].Player == player {
			count++
		}
	}

	// Check diagonals
	if b[0][0].Player == player && b[1][1].Player == player && b[2][2].Player == player {
		count++
	}
	if b[0][2].Player == player && b[1][1].Player == player && b[2][0].Player == player {
		count++
	}

	return count
}
