package main

import (
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"github.com/ivorytoast/replay78/apps"
	"github.com/ivorytoast/replay78/engine"
	"sync"
	"time"

	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true // Allow all origins for development
	},
}

type GameServer struct {
	engine *engine.Engine
	app    *apps.TicTacToeApp
	clients map[*websocket.Conn]bool
	mu     sync.Mutex
}

type Message struct {
	Type    string `json:"type"`
	Payload string `json:"payload"`
}

func NewGameServer() *GameServer {
	e := engine.NewEngine()
	app := apps.NewTicTacToeApp(e)
	e.Register(app)

	e.Run()

	return &GameServer{
		engine:  e,
		app:     app,
		clients: make(map[*websocket.Conn]bool),
	}
}

func (gs *GameServer) handleWebSocket(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Println("Upgrade error:", err)
		return
	}
	defer conn.Close()

	gs.mu.Lock()
	gs.clients[conn] = true
	gs.mu.Unlock()

	// Send initial board state
	gs.sendBoardState(conn)

	for {
		var msg Message
		err := conn.ReadJSON(&msg)
		if err != nil {
			log.Println("Read error:", err)
			gs.mu.Lock()
			delete(gs.clients, conn)
			gs.mu.Unlock()
			break
		}

		gs.handleMessage(conn, msg)
	}
}

func (gs *GameServer) handleMessage(conn *websocket.Conn, msg Message) {
	switch msg.Type {
	case "move":
		gs.engine.In("ttt|move|" + msg.Payload)
		// Give the engine a moment to process
		// In a real implementation, we'd use channels or callbacks
		time.Sleep(50 * time.Millisecond)
		gs.sendBoardState(conn)
	case "endturn":
		gs.engine.In("ttt|endturn|")
		time.Sleep(50 * time.Millisecond)
		gs.sendBoardState(conn)
	case "new":
		gs.engine.In("ttt|new|")
		time.Sleep(50 * time.Millisecond)
		gs.sendBoardState(conn)
	case "show":
		gs.sendBoardState(conn)
	}
}

func (gs *GameServer) sendBoardState(conn *websocket.Conn) {
	state := gs.engine.TTT()
	board := state.GetBoard()

	type CellData struct {
		Player int `json:"player"`
		Power  int `json:"power"`
	}

	boardData := make([][]CellData, 3)
	for i := 0; i < 3; i++ {
		boardData[i] = make([]CellData, 3)
		for j := 0; j < 3; j++ {
			boardData[i][j] = CellData{
				Player: board[i][j].Player,
				Power:  board[i][j].Power,
			}
		}
	}

	// Count lines for each player
	player1Lines := gs.app.CountLines(1)
	player2Lines := gs.app.CountLines(2)

	response := map[string]interface{}{
		"type":             "board_state",
		"board":            boardData,
		"currentPlayer":    state.GetCurrentPlayer(),
		"done":             state.IsDone(),
		"player1PowerBank": state.GetPowerBank(1),
		"player2PowerBank": state.GetPowerBank(2),
		"player1Lines":     player1Lines,
		"player2Lines":     player2Lines,
		"currentPhase":     int(state.GetCurrentPhase()),
		"movementTaken":    state.IsMovementActionTaken(),
	}

	data, _ := json.Marshal(response)
	conn.WriteMessage(websocket.TextMessage, data)
}

func main() {
	gs := NewGameServer()

	http.HandleFunc("/ws", gs.handleWebSocket)
	http.Handle("/", http.FileServer(http.Dir("./web")))

	fmt.Println("Server starting on :8080")
	fmt.Println("Open http://localhost:8080 in your browser")
	log.Fatal(http.ListenAndServe(":8080", nil))
}
