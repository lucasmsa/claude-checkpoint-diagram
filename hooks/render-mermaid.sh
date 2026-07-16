#!/usr/bin/env bash
# Render the last ```mermaid block of a Markdown file (or stdin) as ASCII art.
# Usage: render-mermaid.sh <file.md>   |   ... | render-mermaid.sh
# Prints a hint (not an error) if mermaid-ascii is not installed.
set -uo pipefail

# Resolve the mermaid-ascii binary: PATH first, then common Go bin dirs.
BIN="$(command -v mermaid-ascii 2>/dev/null || true)"
if [ -z "$BIN" ]; then
  for c in "$(go env GOPATH 2>/dev/null)/bin/mermaid-ascii" "$HOME/go/bin/mermaid-ascii"; do
    [ -x "$c" ] && BIN="$c" && break
  done
fi
if [ -z "$BIN" ]; then
  echo "(mermaid-ascii not found; install with: go install github.com/AlexanderGrooff/mermaid-ascii@latest)"
  exit 0
fi

src="${1:--}"
[ "$src" = "-" ] && src=/dev/stdin

# Extract the LAST ```mermaid ... ``` block from the source.
block="$(awk '
  /^```mermaid[[:space:]]*$/ { inblk=1; buf=""; next }
  inblk && /^```[[:space:]]*$/ { inblk=0; last=buf; next }
  inblk { buf = buf $0 "\n" }
  END { printf "%s", last }
' "$src")"

if [ -z "$block" ]; then
  echo "(no mermaid block found)"
  exit 0
fi

printf '%s' "$block" | "$BIN" -y 1 -x 2 -f - 2>&1
