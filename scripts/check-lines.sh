#!/usr/bin/env bash
# Enforces the size rule from docs/adr/0012-agentic-docs-and-line-limits.md:
# agentic docs (*.md) must stay under 100 lines, code under 300 lines.
# PROPOSAL.md (the source PRD) is exempt.
set -euo pipefail

DOC_LIMIT=100
CODE_LIMIT=300
cd "$(git rev-parse --show-toplevel)"

fail=0
check() { # file limit
  local n
  n=$(wc -l < "$1")
  if (( n >= $2 )); then
    echo "FAIL $1: $n lines (limit: under $2)"
    fail=1
  fi
}

while IFS= read -r f; do
  [[ -f "$f" ]] || continue
  case "$f" in
    PROPOSAL.md) ;;
    *.md) check "$f" "$DOC_LIMIT" ;;
    *.go|*.sh|*.yml|*.yaml|*.toml|*.json5|*.qmd|*.typ|*.conf|*Dockerfile)
      check "$f" "$CODE_LIMIT" ;;
  esac
done < <(git ls-files --cached --others --exclude-standard)

if (( fail )); then
  echo "Split the files above; see docs/adr/0012-agentic-docs-and-line-limits.md"
  exit 1
fi
echo "OK: docs < $DOC_LIMIT lines, code < $CODE_LIMIT lines"
