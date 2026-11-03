#!/usr/bin/env bash
# Rejects staged files inside workspace directories. They hold runtime state
# and personal data, so only their top-level .gitignore and README.md are
# ever tracked (ADR 0008, ADR 0024). .caddie/ is never tracked at all.
# Usage: guard-workspace.sh <file>...
set -euo pipefail

DIRS='postings|applied|manual|resumes|cover-letters|handoff|reviews'
WORKSPACE="^(${DIRS}|\\.caddie)/"
ALLOWED="^(${DIRS})/(\\.gitignore|README\\.md)\$"

fail=0
for f in "$@"; do
  if [[ "$f" =~ $WORKSPACE && ! "$f" =~ $ALLOWED ]]; then
    echo "FAIL $f: workspace files are never committed (ADR 0008, ADR 0024)"
    fail=1
  fi
done
exit "$fail"
