package states

type Cell struct {
	Player int
	Power  int
}

type TicTacToeState struct {
	Board                  [][]Cell
	CurrentPlayer          int
	Done                   bool
	Player1PowerBank       int
	Player2PowerBank       int
	Player1FirstTurnDone   bool
	Player2FirstTurnDone   bool
}

func NewTicTacToeState() *TicTacToeState {
	return &TicTacToeState{
		Board: [][]Cell{
			{{Player: 0, Power: 0}, {Player: 0, Power: 0}, {Player: 0, Power: 0}},
			{{Player: 0, Power: 0}, {Player: 0, Power: 0}, {Player: 0, Power: 0}},
			{{Player: 0, Power: 0}, {Player: 0, Power: 0}, {Player: 0, Power: 0}},
		},
		CurrentPlayer:        1,
		Done:                 false,
		Player1PowerBank:     1,
		Player2PowerBank:     1,
		Player1FirstTurnDone: false,
		Player2FirstTurnDone: false,
	}
}

func (tts *TicTacToeState) GetBoard() [][]Cell {
	return tts.Board
}

func (tts *TicTacToeState) GetCurrentPlayer() int {
	return tts.CurrentPlayer
}

func (tts *TicTacToeState) IsDone() bool {
	return tts.Done
}

func (tts *TicTacToeState) SetDone(done bool) {
	tts.Done = done
}

func (tts *TicTacToeState) SetCurrentPlayer(player int) {
	tts.CurrentPlayer = player
}

func (tts *TicTacToeState) SetCell(row, col, player int) {
	tts.Board[row][col] = Cell{Player: player, Power: 1}
}

func (tts *TicTacToeState) IncrementCellPower(row, col int) {
	tts.Board[row][col].Power++
}

func (tts *TicTacToeState) ClearCell(row, col int) {
	tts.Board[row][col] = Cell{Player: 0, Power: 0}
}

func (tts *TicTacToeState) SetCellPower(row, col, power int) {
	tts.Board[row][col].Power = power
}

func (tts *TicTacToeState) GetPowerBank(player int) int {
	if player == 1 {
		return tts.Player1PowerBank
	}
	return tts.Player2PowerBank
}

func (tts *TicTacToeState) SetPowerBank(player int, amount int) {
	if player == 1 {
		tts.Player1PowerBank = amount
	} else {
		tts.Player2PowerBank = amount
	}
}

func (tts *TicTacToeState) IncrementPowerBank(player int, delta int) {
	if player == 1 {
		tts.Player1PowerBank += delta
	} else {
		tts.Player2PowerBank += delta
	}
}

func (tts *TicTacToeState) IsFirstTurnDone(player int) bool {
	if player == 1 {
		return tts.Player1FirstTurnDone
	}
	return tts.Player2FirstTurnDone
}

func (tts *TicTacToeState) SetFirstTurnDone(player int) {
	if player == 1 {
		tts.Player1FirstTurnDone = true
	} else {
		tts.Player2FirstTurnDone = true
	}
}
