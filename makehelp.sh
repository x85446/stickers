#!/usr/bin/env bash
set -euo pipefail

# ============================================================================
# makehelp.sh — Complex logic extracted from Makefile
# ============================================================================
# The Makefile defines WHAT; this script implements HOW for anything too
# complex for inline make recipes (>10 lines, OS branching, multi-step ops).
# ============================================================================

# ---------------------------------------------------------------------------
# Helper functions
# ---------------------------------------------------------------------------

check_command() {
    command -v "$1" &>/dev/null
}

# ---------------------------------------------------------------------------
# Command functions
# ---------------------------------------------------------------------------

cmd_prereqs() {
    echo "Checking prerequisites..."

    if ! check_command go; then
        echo "Error: Go 1.26 or later is required but not installed." >&2
        case "$(uname -s)" in
            Darwin) echo "Install it with: brew install go" >&2 ;;
            *)      echo "Install it from: https://go.dev/dl/" >&2 ;;
        esac
        exit 1
    fi
    echo "  go: $(go env GOVERSION)"

    if ! check_command golangci-lint; then
        echo "  golangci-lint: missing, installing it with go install..."
        go install github.com/golangci/golangci-lint/v2/cmd/golangci-lint@latest
        if ! check_command golangci-lint; then
            echo "Error: golangci-lint is installed in $(go env GOPATH)/bin, which is not on PATH." >&2
            echo "Add it with: export PATH=\"\$PATH:$(go env GOPATH)/bin\"" >&2
            exit 1
        fi
    fi
    echo "  golangci-lint: $(golangci-lint version --short 2>/dev/null || golangci-lint version | head -1)"

    echo "All prerequisites satisfied"
}

# ---------------------------------------------------------------------------
# Dispatcher — one entry per makehelp.sh-delegated target
# ---------------------------------------------------------------------------

case "${1:-}" in
    prereqs) cmd_prereqs ;;
    *)
        echo "Usage: $0 {prereqs}" >&2
        exit 1
        ;;
esac
