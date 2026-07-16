---
name: checkpoint-diagram
description: Append a Mermaid checkpoint diagram summarizing the work done this turn, for review at a pause point. Fires automatically at turn end in auto-accept permission modes (acceptEdits, bypassPermissions), or on demand via /checkpoint-diagram. Skips trivial single-file turns.
---

# Checkpoint diagram

Draw one Mermaid diagram of what changed this turn and append it to the session checkpoint log, so a reviewer catching up on unattended auto-mode work can scan the shape of the changes instead of re-reading the transcript.

## When to draw

Draw when the turn produced substantive work: several file edits, a multi-step change, a new module or flow, or a decision taken among alternatives.

Skip (write nothing, say one line saying so) when the turn was a single small edit, a read-only investigation, or a question answered in prose. A diagram of one edit is noise.

## Pick one diagram type

Choose the type that matches what happened. Do not emit more than one.

| Turn shape | Diagram |
|---|---|
| Multi-step plan or branching decisions | `flowchart` |
| Calls across modules or services, or an agent-and-tool sequence | `sequenceDiagram` |
| Lifecycle or status transitions | `stateDiagram-v2` |
| Which files or modules were touched and how they relate | `flowchart` (dependency graph) |

Default when unsure: `flowchart` of "what I did, then what is next".

## Node rules

- Every node names a concrete action and its target: "Add validate/1 to Portfolio.Import", not "update code".
- Mark decision points. Mark anything skipped or deferred with a dashed node or a `:::deferred` class.
- Cap at about 12 nodes. If the turn was larger, collapse sub-steps into one node and note the count.
- No decorative nodes. If a node carries nothing a reviewer needs, cut it.

## State line (required)

After the diagram, add one line stating what is verified versus assumed. Example: "Compiles; import path tested manually; concurrency path unverified." Never imply more certainty than the turn earned.

## Where to write

Append to `<cwd>/.claude/checkpoints/<YYYY-MM-DD>.md`. Get the date and time from `date`. Add a section:

    ## <n>. <short title>  (<HH:MM>)

    ```mermaid
    <diagram>
    ```

    <two or three line plain summary>. State: <verified vs assumed>.

Number sections sequentially within the file. Ensure `.claude/checkpoints/` is gitignored (add it to the repo `.gitignore` if missing) so checkpoints are never committed.

After appending, print the checkpoint title and the file path so the reviewer knows where to look. A terminal will not render Mermaid, so the file (opened in an IDE preview or on GitHub, both of which render Mermaid natively) is the render surface.

## Worked example

Turn: added portfolio CSV import to the Elixir app.

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

    Added Portfolio.Import.run/1: validates each row, casts valid rows to %Holding{}, bulk-inserts, returns counts plus per-row errors. LiveView upload wiring deferred. State: compiles and unit-tested on the happy path and the malformed-row path; the deferred LiveView path is unbuilt.

## Optional validation

If the Mermaid Chart MCP is available, pass the diagram through `validate_and_render_mermaid_diagram` before writing, to catch syntax errors. Skip if unavailable. Do not make it a hard dependency.
