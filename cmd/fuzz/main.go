package main

import (
	"bufio"
	"flag"
	"fmt"
	"math/rand"
	"os"
	"path/filepath"
	"strings"
	"time"
)

type Move struct {
	row, col int
}

type GameGenerator struct {
	rng *rand.Rand
}

func NewGameGenerator(seed int64) *GameGenerator {
	return &GameGenerator{
		rng: rand.New(rand.NewSource(seed)),
	}
}

// Generate a winning game for specified player on specified row
func (g *GameGenerator) genRowWin(player int, row int) []Move {
	moves := []Move{}

	// Player takes the winning row
	for col := 0; col < 3; col++ {
		moves = append(moves, Move{row, col})
		// Other player moves to a different row
		if col < 2 {
			otherRow := (row + col + 1) % 3
			moves = append(moves, Move{otherRow, col})
		}
	}
	return moves
}

// Generate a winning game for specified player on specified column
func (g *GameGenerator) genColWin(player int, col int) []Move {
	moves := []Move{}

	// Player takes the winning column
	for row := 0; row < 3; row++ {
		moves = append(moves, Move{row, col})
		// Other player moves to a different column
		if row < 2 {
			otherCol := (col + row + 1) % 3
			moves = append(moves, Move{row, otherCol})
		}
	}
	return moves
}

// Generate a diagonal win (0=main diagonal, 1=anti-diagonal)
func (g *GameGenerator) genDiagWin(player int, diagType int) []Move {
	moves := []Move{}

	if diagType == 0 {
		// Main diagonal: (0,0), (1,1), (2,2)
		moves = append(moves, Move{0, 0})
		moves = append(moves, Move{0, 1}) // other
		moves = append(moves, Move{1, 1})
		moves = append(moves, Move{0, 2}) // other
		moves = append(moves, Move{2, 2})
	} else {
		// Anti-diagonal: (0,2), (1,1), (2,0)
		moves = append(moves, Move{0, 2})
		moves = append(moves, Move{0, 0}) // other
		moves = append(moves, Move{1, 1})
		moves = append(moves, Move{0, 1}) // other
		moves = append(moves, Move{2, 0})
	}

	return moves
}

// Generate a tie game
func (g *GameGenerator) genTie() []Move {
	// One specific tie pattern
	return []Move{
		{0, 0}, // X
		{0, 1}, // O
		{0, 2}, // X
		{1, 1}, // O
		{1, 0}, // X
		{2, 0}, // O
		{1, 2}, // X
		{2, 2}, // O
		{2, 1}, // X
	}
}

// Generate a random valid game
func (g *GameGenerator) genRandomGame() []Move {
	board := [3][3]int{}
	moves := []Move{}
	player := 1

	for moveCount := 0; moveCount < 9; moveCount++ {
		// Find available cells
		available := []Move{}
		for r := 0; r < 3; r++ {
			for c := 0; c < 3; c++ {
				if board[r][c] == 0 {
					available = append(available, Move{r, c})
				}
			}
		}

		if len(available) == 0 {
			break
		}

		// Pick random available cell
		move := available[g.rng.Intn(len(available))]
		moves = append(moves, move)
		board[move.row][move.col] = player

		// Check for winner
		if g.checkWinner(board) {
			break
		}

		player = 3 - player
	}

	return moves
}

func (g *GameGenerator) checkWinner(board [3][3]int) bool {
	// Check rows and columns
	for i := 0; i < 3; i++ {
		if board[i][0] != 0 && board[i][0] == board[i][1] && board[i][1] == board[i][2] {
			return true
		}
		if board[0][i] != 0 && board[0][i] == board[1][i] && board[1][i] == board[2][i] {
			return true
		}
	}
	// Check diagonals
	if board[0][0] != 0 && board[0][0] == board[1][1] && board[1][1] == board[2][2] {
		return true
	}
	if board[0][2] != 0 && board[0][2] == board[1][1] && board[1][1] == board[2][0] {
		return true
	}
	return false
}

// Write moves to file with header
func writeMoves(filename, header string, moves []Move, includeShow bool, addInvalidMoves bool, seed int64) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "# Seed: %d\n", seed)
	fmt.Fprintf(f, "# %s\n", header)
	fmt.Fprintln(f, "ttt|new|")

	for _, m := range moves {
		fmt.Fprintf(f, "ttt|move|%d %d\n", m.row, m.col)
	}

	if includeShow {
		fmt.Fprintln(f, "ttt|show|")
	}

	if addInvalidMoves {
		fmt.Fprintln(f, "\n# Try to continue after game ends (should fail)")
		fmt.Fprintln(f, "ttt|move|0 0")
		fmt.Fprintln(f, "ttt|move|1 1")
	}

	fmt.Fprintln(f, "")
	return nil
}

// Generate test with invalid moves
func genInvalidMovesTest(filename string, seed int64) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "# Seed: %d\n", seed)
	fmt.Fprintln(f, "# Invalid Moves Test")
	fmt.Fprintln(f, "ttt|new|")
	fmt.Fprintln(f, "ttt|move|0 0")

	fmt.Fprintln(f, "\n# Try to occupy same cell")
	fmt.Fprintln(f, "ttt|move|0 0")

	fmt.Fprintln(f, "\n# Out of bounds moves")
	fmt.Fprintln(f, "ttt|move|-1 0")
	fmt.Fprintln(f, "ttt|move|0 -1")
	fmt.Fprintln(f, "ttt|move|3 0")
	fmt.Fprintln(f, "ttt|move|0 3")
	fmt.Fprintln(f, "ttt|move|5 5")

	fmt.Fprintln(f, "\n# Continue with valid moves")
	fmt.Fprintln(f, "ttt|move|1 1")
	fmt.Fprintln(f, "ttt|move|0 1")

	fmt.Fprintln(f, "\n# Try occupied cell again")
	fmt.Fprintln(f, "ttt|move|1 1")

	fmt.Fprintln(f, "ttt|show|")
	fmt.Fprintln(f, "")

	return nil
}

// Generate test with multiple games
func genMultipleGamesTest(filename string, gen *GameGenerator, seed int64) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "# Seed: %d\n", seed)
	fmt.Fprintln(f, "# Multiple Games Test")

	// Game 1: Quick X win
	fmt.Fprintln(f, "\n# Game 1: X wins")
	fmt.Fprintln(f, "ttt|new|")
	moves1 := gen.genRowWin(1, 0)
	for _, m := range moves1 {
		fmt.Fprintf(f, "ttt|move|%d %d\n", m.row, m.col)
	}
	fmt.Fprintln(f, "ttt|show|")

	// Game 2: O wins
	fmt.Fprintln(f, "\n# Game 2: O wins")
	fmt.Fprintln(f, "ttt|new|")
	moves2 := gen.genColWin(2, 1)
	for _, m := range moves2 {
		fmt.Fprintf(f, "ttt|move|%d %d\n", m.row, m.col)
	}
	fmt.Fprintln(f, "ttt|show|")

	// Game 3: Tie
	fmt.Fprintln(f, "\n# Game 3: Tie")
	fmt.Fprintln(f, "ttt|new|")
	moves3 := gen.genTie()
	for _, m := range moves3 {
		fmt.Fprintf(f, "ttt|move|%d %d\n", m.row, m.col)
	}
	fmt.Fprintln(f, "ttt|show|")

	fmt.Fprintln(f, "")
	return nil
}

// Generate test with lots of show commands
func genShowCommandsTest(filename string, gen *GameGenerator, seed int64) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "# Seed: %d\n", seed)
	fmt.Fprintln(f, "# Show Commands Test")
	fmt.Fprintln(f, "ttt|new|")
	fmt.Fprintln(f, "ttt|show|")

	moves := gen.genDiagWin(1, 0)
	for i, m := range moves {
		fmt.Fprintf(f, "ttt|move|%d %d\n", m.row, m.col)
		if i%2 == 0 {
			fmt.Fprintln(f, "ttt|show|")
		}
	}

	fmt.Fprintln(f, "ttt|show|")
	fmt.Fprintln(f, "ttt|show|")
	fmt.Fprintln(f, "")

	return nil
}

// Generate edge cases test
func genEdgeCasesTest(filename string, seed int64) error {
	f, err := os.Create(filename)
	if err != nil {
		return err
	}
	defer f.Close()

	fmt.Fprintf(f, "# Seed: %d\n", seed)
	fmt.Fprintln(f, "# Edge Cases Test")

	fmt.Fprintln(f, "\n# Multiple new commands")
	fmt.Fprintln(f, "ttt|new|")
	fmt.Fprintln(f, "ttt|new|")
	fmt.Fprintln(f, "ttt|new|")

	fmt.Fprintln(f, "\n# Show before any moves")
	fmt.Fprintln(f, "ttt|show|")

	fmt.Fprintln(f, "\n# Start a game")
	fmt.Fprintln(f, "ttt|move|1 1")
	fmt.Fprintln(f, "ttt|move|0 0")

	fmt.Fprintln(f, "\n# New game in the middle")
	fmt.Fprintln(f, "ttt|new|")
	fmt.Fprintln(f, "ttt|show|")

	fmt.Fprintln(f, "\n# Complete a game")
	fmt.Fprintln(f, "ttt|move|0 0")
	fmt.Fprintln(f, "ttt|move|1 0")
	fmt.Fprintln(f, "ttt|move|0 1")
	fmt.Fprintln(f, "ttt|move|1 1")
	fmt.Fprintln(f, "ttt|move|0 2")
	fmt.Fprintln(f, "ttt|show|")

	fmt.Fprintln(f, "")
	return nil
}

type TestConfig struct {
	outputDir      string
	numRandom      int
	seed           int64
	generateXWins  bool
	generateOWins  bool
	generateTies   bool
	generateInvalid bool
	generateMulti  bool
	generateShow   bool
	generateEdge   bool
}

func interactiveMode() TestConfig {
	config := TestConfig{
		outputDir: "fuzz_tests",
	}
	scanner := bufio.NewScanner(os.Stdin)

	fmt.Println("=== Fuzz Test Generator ===")
	fmt.Println()

	// Number of random tests
	fmt.Print("Number of random tests (default 10): ")
	scanner.Scan()
	input := strings.TrimSpace(scanner.Text())
	if input == "" {
		config.numRandom = 10
	} else {
		fmt.Sscanf(input, "%d", &config.numRandom)
	}

	// Seed selection
	fmt.Print("Use specific seed? (y/N): ")
	scanner.Scan()
	useSeed := strings.TrimSpace(scanner.Text())
	if useSeed == "y" || useSeed == "Y" {
		fmt.Print("Enter seed: ")
		scanner.Scan()
		fmt.Sscanf(strings.TrimSpace(scanner.Text()), "%d", &config.seed)
	} else {
		config.seed = time.Now().UnixNano()
	}

	fmt.Println()
	fmt.Println("Select test types to generate (press Enter to select all):")
	fmt.Println("  [1] X wins (rows, cols, diagonals)")
	fmt.Println("  [2] O wins (rows, cols, diagonals)")
	fmt.Println("  [3] Tie games")
	fmt.Println("  [4] Invalid moves")
	fmt.Println("  [5] Multiple games")
	fmt.Println("  [6] Show commands")
	fmt.Println("  [7] Edge cases")
	fmt.Print("Enter numbers separated by spaces (e.g., 1 2 3) or press Enter for all: ")

	scanner.Scan()
	selection := strings.TrimSpace(scanner.Text())

	if selection == "" {
		// Select all
		config.generateXWins = true
		config.generateOWins = true
		config.generateTies = true
		config.generateInvalid = true
		config.generateMulti = true
		config.generateShow = true
		config.generateEdge = true
	} else {
		// Parse selections
		for _, char := range selection {
			switch char {
			case '1':
				config.generateXWins = true
			case '2':
				config.generateOWins = true
			case '3':
				config.generateTies = true
			case '4':
				config.generateInvalid = true
			case '5':
				config.generateMulti = true
			case '6':
				config.generateShow = true
			case '7':
				config.generateEdge = true
			}
		}
	}

	fmt.Println()
	return config
}

func main() {
	// Define flags
	outputDir := flag.String("output", "fuzz_tests", "Output directory for generated test files")
	numRandom := flag.Int("random", 10, "Number of random test games to generate")
	seed := flag.Int64("seed", 0, "Random seed (0 for random)")
	flag.Parse()

	var config TestConfig

	// Check if running in flag mode (any flag provided) or interactive mode
	if flag.NFlag() > 0 {
		// Flag mode
		config.outputDir = *outputDir
		config.numRandom = *numRandom
		if *seed == 0 {
			config.seed = time.Now().UnixNano()
		} else {
			config.seed = *seed
		}
		// Generate all types in flag mode
		config.generateXWins = true
		config.generateOWins = true
		config.generateTies = true
		config.generateInvalid = true
		config.generateMulti = true
		config.generateShow = true
		config.generateEdge = true
	} else {
		// Interactive mode
		config = interactiveMode()
	}

	// Create output directory
	if err := os.MkdirAll(config.outputDir, 0755); err != nil {
		fmt.Printf("Error creating output directory: %v\n", err)
		os.Exit(1)
	}

	gen := NewGameGenerator(config.seed)
	fmt.Printf("Generating fuzz tests with seed %d...\n", config.seed)

	testCount := 0
	var categories []string

	// X wins
	if config.generateXWins {
		for row := 0; row < 3; row++ {
			filename := filepath.Join(config.outputDir, fmt.Sprintf("test_%03d_x_wins_row_%d.txt", testCount, row))
			moves := gen.genRowWin(1, row)
			if err := writeMoves(filename, fmt.Sprintf("X Wins - Row %d", row), moves, true, true, config.seed); err != nil {
				fmt.Printf("Error writing %s: %v\n", filename, err)
			} else {
				testCount++
			}
		}

		for col := 0; col < 3; col++ {
			filename := filepath.Join(config.outputDir, fmt.Sprintf("test_%03d_x_wins_col_%d.txt", testCount, col))
			moves := gen.genColWin(1, col)
			if err := writeMoves(filename, fmt.Sprintf("X Wins - Column %d", col), moves, true, true, config.seed); err != nil {
				fmt.Printf("Error writing %s: %v\n", filename, err)
			} else {
				testCount++
			}
		}

		for diag := 0; diag < 2; diag++ {
			diagName := "Main"
			if diag == 1 {
				diagName = "Anti"
			}
			filename := filepath.Join(config.outputDir, fmt.Sprintf("test_%03d_x_wins_diag_%s.txt", testCount, diagName))
			moves := gen.genDiagWin(1, diag)
			if err := writeMoves(filename, fmt.Sprintf("X Wins - %s Diagonal", diagName), moves, true, true, config.seed); err != nil {
				fmt.Printf("Error writing %s: %v\n", filename, err)
			} else {
				testCount++
			}
		}
		categories = append(categories, "X wins: 8 tests")
	}

	// O wins
	if config.generateOWins {
		for row := 0; row < 3; row++ {
			filename := filepath.Join(config.outputDir, fmt.Sprintf("test_%03d_o_wins_row_%d.txt", testCount, row))
			moves := gen.genRowWin(2, row)
			if err := writeMoves(filename, fmt.Sprintf("O Wins - Row %d", row), moves, true, true, config.seed); err != nil {
				fmt.Printf("Error writing %s: %v\n", filename, err)
			} else {
				testCount++
			}
		}

		for col := 0; col < 3; col++ {
			filename := filepath.Join(config.outputDir, fmt.Sprintf("test_%03d_o_wins_col_%d.txt", testCount, col))
			moves := gen.genColWin(2, col)
			if err := writeMoves(filename, fmt.Sprintf("O Wins - Column %d", col), moves, true, true, config.seed); err != nil {
				fmt.Printf("Error writing %s: %v\n", filename, err)
			} else {
				testCount++
			}
		}

		for diag := 0; diag < 2; diag++ {
			diagName := "Main"
			if diag == 1 {
				diagName = "Anti"
			}
			filename := filepath.Join(config.outputDir, fmt.Sprintf("test_%03d_o_wins_diag_%s.txt", testCount, diagName))
			moves := gen.genDiagWin(2, diag)
			if err := writeMoves(filename, fmt.Sprintf("O Wins - %s Diagonal", diagName), moves, true, true, config.seed); err != nil {
				fmt.Printf("Error writing %s: %v\n", filename, err)
			} else {
				testCount++
			}
		}
		categories = append(categories, "O wins: 8 tests")
	}

	// Tie game
	if config.generateTies {
		filename := filepath.Join(config.outputDir, fmt.Sprintf("test_%03d_tie.txt", testCount))
		moves := gen.genTie()
		if err := writeMoves(filename, "Tie Game", moves, true, true, config.seed); err != nil {
			fmt.Printf("Error writing %s: %v\n", filename, err)
		} else {
			testCount++
		}
		categories = append(categories, "Tie game: 1 test")
	}

	// Invalid moves test
	if config.generateInvalid {
		filename := filepath.Join(config.outputDir, fmt.Sprintf("test_%03d_invalid_moves.txt", testCount))
		if err := genInvalidMovesTest(filename, config.seed); err != nil {
			fmt.Printf("Error writing %s: %v\n", filename, err)
		} else {
			testCount++
		}
		categories = append(categories, "Invalid moves: 1 test")
	}

	// Multiple games test
	if config.generateMulti {
		filename := filepath.Join(config.outputDir, fmt.Sprintf("test_%03d_multiple_games.txt", testCount))
		if err := genMultipleGamesTest(filename, gen, config.seed); err != nil {
			fmt.Printf("Error writing %s: %v\n", filename, err)
		} else {
			testCount++
		}
		categories = append(categories, "Multiple games: 1 test")
	}

	// Show commands test
	if config.generateShow {
		filename := filepath.Join(config.outputDir, fmt.Sprintf("test_%03d_show_commands.txt", testCount))
		if err := genShowCommandsTest(filename, gen, config.seed); err != nil {
			fmt.Printf("Error writing %s: %v\n", filename, err)
		} else {
			testCount++
		}
		categories = append(categories, "Show commands: 1 test")
	}

	// Edge cases test
	if config.generateEdge {
		filename := filepath.Join(config.outputDir, fmt.Sprintf("test_%03d_edge_cases.txt", testCount))
		if err := genEdgeCasesTest(filename, config.seed); err != nil {
			fmt.Printf("Error writing %s: %v\n", filename, err)
		} else {
			testCount++
		}
		categories = append(categories, "Edge cases: 1 test")
	}

	// Random games
	if config.numRandom > 0 {
		for i := 0; i < config.numRandom; i++ {
			filename := filepath.Join(config.outputDir, fmt.Sprintf("test_%03d_random_%d.txt", testCount, i))
			moves := gen.genRandomGame()
			if err := writeMoves(filename, fmt.Sprintf("Random Game %d", i), moves, true, false, config.seed); err != nil {
				fmt.Printf("Error writing %s: %v\n", filename, err)
			} else {
				testCount++
			}
		}
		categories = append(categories, fmt.Sprintf("Random games: %d tests", config.numRandom))
	}

	fmt.Printf("\n✅ Generated %d test files in %s/\n", testCount, config.outputDir)
	if len(categories) > 0 {
		fmt.Println("\nTest categories:")
		for _, cat := range categories {
			fmt.Printf("  - %s\n", cat)
		}
	}
	fmt.Printf("\nTotal: %d tests\n", testCount)

	fmt.Println("\nTo run a test:")
	fmt.Printf("  go run main.go %s/test_000_*.txt\n", config.outputDir)
	fmt.Println("\nTo run all tests:")
	fmt.Printf("  go run main.go --regression\n")
}
