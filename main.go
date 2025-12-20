package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"github.com/ivorytoast/replay78/apps"
	"github.com/ivorytoast/replay78/engine"
	"strings"
	"time"
)

func main() {
	regression := flag.Bool("regression", false, "Run regression tests")
	flag.Parse()

	if *regression {
		regressionTestFiles := discoverFuzzTestBaselines()
		runRegressionTests(regressionTestFiles)
		return
	}

	l := engine.NewEngine()

	app := apps.NewTicTacToeApp(l)
	l.Register(app)

	go l.Run()

	args := flag.Args()
	if len(args) > 0 {
		replayFromFile(l, args[0])
	} else {
		interactiveMode(l)
	}
}

func discoverFuzzTestBaselines() []string {
	baselineDir := "fuzz_baselines"
	fuzzTestDir := "fuzz_tests"

	// Check if fuzz_tests directory exists
	if _, err := os.Stat(fuzzTestDir); os.IsNotExist(err) {
		return []string{}
	}

	// Create baseline directory if it doesn't exist
	os.MkdirAll(baselineDir, 0755)

	// Find all .txt files in fuzz_tests
	testFiles, err := filepath.Glob(filepath.Join(fuzzTestDir, "*.txt"))
	if err != nil {
		fmt.Printf("Error reading fuzz tests: %v\n", err)
		return []string{}
	}

	var baselineFiles []string

	for _, testFile := range testFiles {
		// Generate baseline log filename
		baseName := strings.TrimSuffix(filepath.Base(testFile), ".txt")
		baselineLog := filepath.Join(baselineDir, baseName+".log")

		// If baseline doesn't exist, create it
		if _, err := os.Stat(baselineLog); os.IsNotExist(err) {
			fmt.Printf("Creating baseline: %s from %s\n", baselineLog, testFile)

			// Create engine with baseline log file
			l := engine.NewEngineWithLogFile(baselineLog)

			app := apps.NewTicTacToeApp(l)
			l.Register(app)

			go l.Run()

			// Read and replay test file
			file, err := os.Open(testFile)
			if err != nil {
				fmt.Printf("Error opening test file %s: %v\n", testFile, err)
				continue
			}

			scanner := bufio.NewScanner(file)
			for scanner.Scan() {
				line := strings.TrimSpace(scanner.Text())
				if line == "" || strings.HasPrefix(line, "#") {
					continue
				}
				l.In(line)
			}
			file.Close()
		}

		baselineFiles = append(baselineFiles, baselineLog)
	}

	return baselineFiles
}

func runRegressionTests(logFiles []string) {
	if len(logFiles) == 0 {
		return
	}

	fmt.Println("=== Running Regression Tests ===")
	fmt.Println()
	allPassed := true

	// Store replay log names for comparison
	type testPair struct {
		original string
		replay   string
	}
	var testPairs []testPair

	// Redirect stdout to suppress game output
	oldStdout := os.Stdout
	devNull, _ := os.Open(os.DevNull)
	os.Stdout = devNull

	// First pass: Create all replay logs
	for _, logFile := range logFiles {
		// Skip if file doesn't exist
		if _, err := os.Stat(logFile); os.IsNotExist(err) {
			continue
		}

		// Parse inputs from original log
		inputs, err := parseLogForInputs(logFile)
		if err != nil {
			// Restore stdout for error message
			os.Stdout = oldStdout
			fmt.Printf("  ❌ FAILED - Error reading log: %v\n", err)
			os.Stdout = devNull
			allPassed = false
			continue
		}

		// Create replay log filename - find next available number
		baseName := strings.TrimSuffix(logFile, ".log")
		replayBase := baseName + "-replay"

		// Find the highest numbered replay file
		highestNum := 0
		for i := 1; i < 1000; i++ {
			testFile := fmt.Sprintf("%s-%d.log", replayBase, i)
			if _, err := os.Stat(testFile); err == nil {
				highestNum = i
			} else {
				break
			}
		}

		replayLogName := fmt.Sprintf("%s-%d.log", replayBase, highestNum+1)

		// Create new engine with custom log file
		l := engine.NewEngineWithLogFile(replayLogName)

		app := apps.NewTicTacToeApp(l)
		l.Register(app)

		go l.Run()

		// Replay inputs
		for _, input := range inputs {
			l.In(input)
		}

		testPairs = append(testPairs, testPair{logFile, replayLogName})
	}

	// Wait once for all engines to finish writing
	time.Sleep(500 * time.Millisecond)

	// Restore stdout
	os.Stdout = oldStdout
	devNull.Close()

	// Second pass: Compare all logs
	for _, pair := range testPairs {
		match, err := compareLogFiles(pair.original, pair.replay)
		if err != nil {
			fmt.Printf("  ❌ FAILED - Error comparing %s: %v\n", pair.original, err)
			allPassed = false
			continue
		}

		if match {
			fmt.Printf("  ✅ PASSED - %s\n", filepath.Base(pair.original))
		} else {
			fmt.Printf("  ❌ FAILED - %s (output differs from original)\n", filepath.Base(pair.original))
			allPassed = false
		}
	}

	fmt.Println()
	fmt.Println("================================")
	if allPassed {
		fmt.Println("✅ All regression tests PASSED")
	} else {
		fmt.Println("❌ Some regression tests FAILED")
	}
	fmt.Println()
}

func replayFromFile(l *engine.Engine, filename string) {
	file, err := os.Open(filename)
	if err != nil {
		fmt.Printf("Error opening file: %v\n", err)
		return
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		fmt.Printf("Processing: %s\n", line)
		l.In(line)
	}

	if err := scanner.Err(); err != nil {
		fmt.Printf("Error reading file: %v\n", err)
	}
}

func interactiveMode(l *engine.Engine) {
	fmt.Println("=== Tic Tac Toe ===")
	for {
		fmt.Println("\nOptions: [n]ew game, [s]how board, [m]ove, [q]uit")
		fmt.Print("-> ")
		var choice string
		fmt.Scanln(&choice)
		switch choice {
		case "n", "new":
			l.In("ttt|new|")
		case "s", "show":
			l.In("ttt|show|")
		case "m", "move":
			var fromRow, fromCol, toRow, toCol int
			fmt.Print("Enter from row (0-2): ")
			fmt.Scanln(&fromRow)
			fmt.Print("Enter from col (0-2): ")
			fmt.Scanln(&fromCol)
			fmt.Print("Enter to row (0-2): ")
			fmt.Scanln(&toRow)
			fmt.Print("Enter to col (0-2): ")
			fmt.Scanln(&toCol)
			l.In(fmt.Sprintf("ttt|move|%d %d %d %d", fromRow, fromCol, toRow, toCol))
		case "q", "quit":
			return
		default:
			fmt.Println("Invalid option")
		}
	}
}

func parseLogForInputs(filename string) ([]string, error) {
	file, err := os.Open(filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var inputs []string
	scanner := bufio.NewScanner(file)
	for scanner.Scan() {
		line := scanner.Text()
		parts := strings.SplitN(line, "|", 3)
		if len(parts) >= 3 && parts[1] == "I" {
			inputs = append(inputs, parts[2])
		}
	}
	return inputs, scanner.Err()
}

func compareLogFiles(originalFile, replayFile string) (bool, error) {
	original, err := os.ReadFile(originalFile)
	if err != nil {
		return false, err
	}
	replay, err := os.ReadFile(replayFile)
	if err != nil {
		return false, err
	}

	origLines := strings.Split(string(original), "\n")
	replayLines := strings.Split(string(replay), "\n")

	if len(origLines) != len(replayLines) {
		return false, nil
	}

	for i := range origLines {
		origParts := strings.SplitN(origLines[i], "|", 2)
		replayParts := strings.SplitN(replayLines[i], "|", 2)

		if len(origParts) < 2 || len(replayParts) < 2 {
			if origLines[i] != replayLines[i] {
				return false, nil
			}
			continue
		}

		if origParts[1] != replayParts[1] {
			return false, nil
		}
	}

	return true, nil
}
