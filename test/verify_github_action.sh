#!/bin/bash

# Configuration
SCRIPT_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"
WORKFLOW_FILE="$SCRIPT_DIR/../.github/workflows/backend.yml"
BACKEND_DIR="hr-backend"

echo "Running meta-tests for GitHub Action..."

# 1. Check if file exists
if [ ! -f "$WORKFLOW_FILE" ]; then
    echo "FAILED: Workflow file $WORKFLOW_FILE does not exist"
    exit 1
fi

# 2. Check for triggers
grep -q "push:" "$WORKFLOW_FILE" || { echo "FAILED: Push trigger missing"; exit 1; }
grep -q "pull_request:" "$WORKFLOW_FILE" || { echo "FAILED: PR trigger missing"; exit 1; }
grep -q "main" "$WORKFLOW_FILE" || { echo "FAILED: Branch 'main' missing in triggers"; exit 1; }

# 3. Check for working-directory
if [ -f "$SCRIPT_DIR/../go.mod" ]; then
    # We are in the backend root directory (or its subdirectory)
    grep -q "working-directory: $BACKEND_DIR" "$WORKFLOW_FILE" && { echo "FAILED: Redundant working-directory ($BACKEND_DIR) set in backend root"; exit 1; }
else
    # We are in a parent directory (like hr-system)
    grep -q "working-directory: $BACKEND_DIR" "$WORKFLOW_FILE" || { echo "FAILED: Correct working-directory ($BACKEND_DIR) not set"; exit 1; }
fi

# 4. Check for essential steps
grep -q "extractions/setup-just" "$WORKFLOW_FILE" || { echo "FAILED: setup-just step missing"; exit 1; }
grep -q "just lint" "$WORKFLOW_FILE" || { echo "FAILED: just lint step missing"; exit 1; }
grep -q "just test" "$WORKFLOW_FILE" || { echo "FAILED: just test step missing"; exit 1; }
grep -q "just build" "$WORKFLOW_FILE" || { echo "FAILED: just build step missing"; exit 1; }

echo "PASSED: GitHub Action meta-tests"
exit 0
