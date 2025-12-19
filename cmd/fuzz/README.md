# Fuzz - Tic-Tac-Toe Test Generator

Standalone tool to generate comprehensive test cases for the tic-tac-toe replay system.

## Installation

### Build Binary
```bash
go build -o fuzz cmd/fuzz/main.go
```

### Run Directly
```bash
go run cmd/fuzz/main.go [options]
```

## Usage

### Basic Usage
```bash
# Generate default test suite (21 deterministic + 10 random tests)
./fuzz

# OR
go run cmd/fuzz/main.go
```

### Options

```
-output <dir>    Output directory for test files (default: "fuzz_tests")
-random <n>      Number of random test games (default: 10)
-seed <n>        Random seed for reproducibility (default: current timestamp)
```

### Examples

```bash
# Generate 20 random tests
./fuzz -random 20

# Use custom output directory
./fuzz -output my_tests

# Use specific seed for reproducibility
./fuzz -seed 12345

# Combine options
./fuzz -output custom_tests -random 50 -seed 42
```

## Generated Tests

### Deterministic Tests (21 files)

1. **X Wins (8 tests)**
   - 3 row wins (one per row)
   - 3 column wins (one per column)
   - 2 diagonal wins (main and anti-diagonal)

2. **O Wins (8 tests)**
   - 3 row wins, 3 column wins, 2 diagonal wins

3. **Tie Game (1 test)**

4. **Invalid Moves (1 test)**
   - Occupied cells
   - Out of bounds moves
   - Mixed with valid moves

5. **Multiple Games (1 test)**
   - Three games in sequence

6. **Show Commands (1 test)**
   - Multiple show commands

7. **Edge Cases (1 test)**
   - Multiple new commands
   - New in middle of game

### Random Tests (configurable)

Randomly generated valid games with unpredictable move sequences.

## Output Format

Each test file includes:
- Seed ID (first line comment)
- Test description
- Commands in `topic|action|payload` format

Example:
```
# Seed: 12345
# X Wins - Row 0
ttt|new|
ttt|move|0 0
ttt|move|1 0
...
```

## Integration

### Run Tests with Main Program
```bash
# Single test
go run main.go fuzz_tests/test_000_x_wins_row_0.txt

# All tests with regression mode
go run main.go --regression
```

### Build and Install
```bash
# Build binary
go build -o fuzz cmd/fuzz/main.go

# Move to PATH (optional)
mv fuzz /usr/local/bin/

# Run from anywhere
fuzz -random 100 -output ~/my-tests
```

## Reproducibility

The seed is stored in every test file, allowing exact reproduction:

```bash
# Generate tests with seed 999
./fuzz -seed 999

# Later, regenerate the exact same tests
./fuzz -seed 999
```

## Use Cases

- **Initial Setup**: Generate baseline test suite
- **Regression Testing**: Create tests that run automatically
- **Stress Testing**: Generate hundreds of random scenarios
- **Bug Reproduction**: Use specific seeds to recreate failures
- **CI/CD**: Generate fresh tests for each build
