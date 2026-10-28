#!/usr/bin/env bash
# Rejects staged files inside workspace directories. They hold runtime state
# and personal data, so only their .gitignore is ever tracked (ADR 0008).
# Usage: guard-workspace.sh <file>...
set -euo pipefail

WORKSPACE='^(postings|applied|manual|resumes|cover-letters|handoff|reviews|\.caddie)/'

fail=0
for f in "$@"; do
  if [[ "$f" =~ $WORKSPACE && "$(basename "$f")" != .gitignore ]]; then
    echo "FAIL $f: workspace files are never committed (ADR 0008)"
    fail=1
  fi
done
exit "$fail"
