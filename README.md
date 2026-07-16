# checkpoint-diagram

A Claude Code plugin that draws one Mermaid diagram of what changed each time Claude pauses in an auto-accept mode, and appends it to a per-day checkpoint log. When you let the agent run unattended, you get a visual catch-up instead of re-reading the transcript.

It does nothing in normal or plan mode. It only fires when you are in an auto-accept permission mode (`acceptEdits`, `bypassPermissions`, `auto`), which is exactly when you are not watching each step.

## How it works

1. A `Stop` hook runs at turn end. If the permission mode is not auto-accept, it exits and does nothing.
2. Trivial turns are skipped: a short final message with a clean working tree produces no diagram.
3. Otherwise the hook blocks the stop once and asks Claude to invoke the `checkpoint-diagram` skill.
4. The skill picks one diagram type (flowchart, sequence, state, or dependency graph) based on what the turn did, then appends a numbered section to `.claude/checkpoints/<YYYY-MM-DD>.md`.
5. A sentinel file plus the `stop_hook_active` flag prevent any re-trigger loop.

A terminal does not render Mermaid, so the checkpoint file is the render surface. Open it in an IDE Markdown preview or on GitHub, both of which render Mermaid natively.

## Example checkpoint

````markdown
## 1. CSV import path  (14:22)

```mermaid
flowchart TD
    A[Add Portfolio.Import.run/1] --> B{Row valid?}
    B -->|yes| C[Cast to %Holding{}]
    B -->|no| D[Collect error, skip row]
    C --> E[Bulk insert via Repo.insert_all]
    D --> F[Return {:ok, inserted, errors}]
    E --> F
    G[Wire to ImportLive upload]:::deferred
    F -.-> G
    classDef deferred stroke-dasharray: 4 4;
```

Added Portfolio.Import.run/1: validates each row, casts valid rows, bulk-inserts,
returns counts plus per-row errors. LiveView wiring deferred.
State: compiles and unit-tested on happy and malformed-row paths; deferred path unbuilt.
````

## Install

### As a plugin

```
/plugin marketplace add lucasmsa/claude-checkpoint-diagram
/plugin install checkpoint-diagram@lucasmsa-plugins
```

The plugin ships both the skill and the `Stop` hook, so no manual settings edit is needed.

### Manual

1. Copy `skills/checkpoint-diagram/` to `~/.claude/skills/checkpoint-diagram/`.
2. Copy `hooks/checkpoint-diagram.sh` to `~/.claude/hooks/` and `chmod +x` it.
3. Add the hook to the `Stop` array in `~/.claude/settings.json`:

```json
{
  "hooks": {
    "Stop": [
      {
        "hooks": [
          {
            "type": "command",
            "command": "/absolute/path/to/.claude/hooks/checkpoint-diagram.sh",
            "timeout": 60
          }
        ]
      }
    ]
  }
}
```

## Configuration

- **Which modes trigger it.** Edit the `case` statement in `hooks/checkpoint-diagram.sh`. Default set: `acceptEdits`, `bypassPermissions`, `auto`, `dontAsk`.
- **What counts as trivial.** The hook skips when the final message is under 400 characters and the working tree has no changes. Adjust the length threshold in the script.
- **Where checkpoints land.** `.claude/checkpoints/<YYYY-MM-DD>.md` in the working directory. Add `.claude/checkpoints/` to your `.gitignore`.

## Requirements

- `jq` (the hook parses the hook payload with it).
- `git` is optional; if present, a clean working tree is used as one signal that a turn was trivial.

## Manual use

Run `/checkpoint-diagram` at any time to draw a checkpoint for the current turn, regardless of permission mode.

## License

MIT
