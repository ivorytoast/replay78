#!/bin/bash
# Claude Code Hook: Post-Edit
# Runs regression tests after code changes to Go files

set -e

# Get the file that was edited from the hook arguments
EDITED_FILE="${HOOK_FILE_PATH:-}"

# Only run tests if a Go file was edited
if [[ "$EDITED_FILE" == *.go ]]; then
    echo "🔍 Code change detected in $EDITED_FILE"
    echo "🧪 Running regression tests..."
    echo ""

    # Run regression tests
    if go run main.go --regression 2>&1; then
        echo ""
        echo "✅ All regression tests passed! Safe to continue."
        exit 0
    else
        echo ""
        echo "❌ REGRESSION TESTS FAILED!"
        echo "⚠️  Please fix the failing tests before proceeding."
        exit 1
    fi
fi

# If not a Go file, exit successfully without running tests
exit 0
