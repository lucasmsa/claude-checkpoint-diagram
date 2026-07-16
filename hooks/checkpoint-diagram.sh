#!/usr/bin/env bash
# Stop hook: in auto-accept permission modes, make Claude append a checkpoint
# diagram before the turn ends. Loop-safe via stop_hook_active + a sentinel file.
set -euo pipefail

input=$(cat)

mode=$(printf '%s' "$input"      | jq -r '.permission_mode // "default"')
stop_active=$(printf '%s' "$input" | jq -r '.stop_hook_active // false')
session=$(printf '%s' "$input"   | jq -r '.session_id // "nosession"')
cwd=$(printf '%s' "$input"       | jq -r '.cwd // "."')
last_msg=$(printf '%s' "$input"  | jq -r '.last_assistant_message // ""')

# Only in auto-accept modes. Edit this set if your build names them differently.
case "$mode" in
  acceptEdits|bypassPermissions|auto|dontAsk) ;;
  *) exit 0 ;;
esac

guard="${TMPDIR:-/tmp}/cc-checkpoint-${session}.pending"

# This Stop is the re-trigger after we already asked for a diagram: let it through.
if [[ "$stop_active" == "true" || -f "$guard" ]]; then
  rm -f "$guard"
  exit 0
fi

# Skip trivial turns: short final message and no working-tree changes.
changed=""
if git -C "$cwd" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
  changed=$(git -C "$cwd" status --porcelain 2>/dev/null | head -c 1)
fi
if [[ -z "$changed" && ${#last_msg} -lt 400 ]]; then
  exit 0
fi

# Block the stop and ask Claude to draw the checkpoint, then stop.
touch "$guard"
cat <<'JSON'
{"decision":"block","reason":"Before stopping: you are in an auto-accept mode, so invoke the checkpoint-diagram skill now. Draw one Mermaid diagram of the substantive work from this turn and append it to .claude/checkpoints/<today>.md per the skill's rules, then stop. If the turn was genuinely trivial, skip the diagram and say so in one line."}
JSON
exit 0
