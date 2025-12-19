# Claude Code Hooks

This directory contains hooks that automatically run during Claude Code operations.

## Available Hooks

### `post-edit.sh`

**Trigger**: Runs automatically after Claude edits any Go file
**Purpose**: Ensures all regression tests pass after code changes

**Behavior**:
- Detects when a `.go` file is edited
- Automatically runs `go run main.go --regression`
- Reports test results
- Blocks if tests fail (exit code 1)
- Allows continuation if tests pass (exit code 0)

**Example Output**:
```
🔍 Code change detected in apps/tictactoe.go
🧪 Running regression tests...

=== Running Regression Tests ===

  ✅ PASSED - test_000_x_wins_row_0.log
  ...
  ✅ PASSED - test_022_random_1.log

================================
✅ All regression tests PASSED

✅ All regression tests passed! Safe to continue.
```

## How It Works

1. Claude Code detects when a file is edited
2. The hook checks if it's a `.go` file
3. If yes, runs the full regression test suite (23 tests)
4. Reports pass/fail status
5. Claude can see the results and take appropriate action

## Disabling Hooks

To temporarily disable the hook:
```bash
# Rename the hook file
mv .claude/hooks/post-edit.sh .claude/hooks/post-edit.sh.disabled
```

To re-enable:
```bash
# Rename it back
mv .claude/hooks/post-edit.sh.disabled .claude/hooks/post-edit.sh
```

## Benefits

- ✅ **Automatic validation**: Never accidentally break the code
- ✅ **Fast feedback**: Know immediately if a change breaks tests
- ✅ **Confidence**: Every edit is validated against all 23 test scenarios
- ✅ **Safety net**: Catches regressions before they propagate

## Customization

You can modify the hook to:
- Run on different file types (change `*.go` pattern)
- Run different commands
- Add additional checks (linting, formatting, etc.)
- Send notifications
- Generate reports
