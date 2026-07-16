#!/usr/bin/env bash
# Stop hook: in auto-accept permission modes, ask Claude to append a checkpoint
# diagram before the turn ends, but only when the turn actually did work.
# A pure-conversation turn (no tool calls) is left completely silent: no block,
# no reprompt, no message. Loop-safe via stop_hook_active + a sentinel file.
set -uo pipefail

input=$(cat)

mode=$(printf '%s' "$input"        | jq -r '.permission_mode // "default"')
stop_active=$(printf '%s' "$input" | jq -r '.stop_hook_active // false')
session=$(printf '%s' "$input"     | jq -r '.session_id // "nosession"')
cwd=$(printf '%s' "$input"         | jq -r '.cwd // "."')
transcript=$(printf '%s' "$input"  | jq -r '.transcript_path // ""')
last_msg=$(printf '%s' "$input"    | jq -r '.last_assistant_message // ""')

# Only in auto-accept modes. Edit this set if your build names them differently.
case "$mode" in
  acceptEdits|bypassPermissions|auto|dontAsk) ;;
  *) exit 0 ;;
esac

guard="${TMPDIR:-/tmp}/cc-checkpoint-${session}.pending"

# Re-trigger after we already asked for a diagram: let it through.
if [[ "$stop_active" == "true" || -f "$guard" ]]; then
  rm -f "$guard"
  exit 0
fi

# Did this turn use any tools? A pure-conversation turn has nothing to draw, so
# stay silent. Prefer the transcript; if it cannot be read, fall back to a
# message-length + working-tree heuristic.
markers=""
if [[ -n "$transcript" && -r "$transcript" ]]; then
  markers=$(tail -n 400 "$transcript" | jq -rc '
      if .type=="user" then
        (if (.message.content|type=="string") then "U"
         elif (.message.content|type=="array") and (any(.message.content[]?; .type=="text")) then "U"
         else "R" end)
      elif .type=="assistant" then
        ("A:" + ([.message.content[]? | select(.type=="tool_use") | .name] | join(",")))
      else "-" end' 2>/dev/null)
fi

if [[ -n "$markers" ]]; then
  # Transcript readable: skip silently if no tool_use since the last user message.
  if ! printf '%s\n' "$markers" | awk '/^U$/{u=0} /^A:/{ if (length($0)>2) u=1 } END{ exit (u+0)?0:1 }'; then
    exit 0
  fi
else
  # No transcript available: skip only clearly-trivial turns.
  changed=""
  if git -C "$cwd" rev-parse --is-inside-work-tree >/dev/null 2>&1; then
    changed=$(git -C "$cwd" status --porcelain 2>/dev/null | head -c 1)
  fi
  if [[ -z "$changed" && ${#last_msg} -lt 400 ]]; then
    exit 0
  fi
fi

# Block the stop and ask Claude to draw the checkpoint (or skip silently if trivial).
touch "$guard"
cat <<'JSON'
{"decision":"block","reason":"Before stopping: you are in an auto-accept mode, so invoke the checkpoint-diagram skill now. Draw one Mermaid diagram of the substantive work from this turn and append it to .claude/checkpoints/<today>.md per the skill's rules, then stop. If the turn was genuinely trivial with nothing to show, skip silently: do not draw, and do not say anything about checkpoints or skipping."}
JSON
exit 0
