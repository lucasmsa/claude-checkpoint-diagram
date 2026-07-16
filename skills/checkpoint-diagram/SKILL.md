---
name: checkpoint-diagram
description: At a pause, append a Mermaid checkpoint of this turn's work to .claude/checkpoints/<date>.md AND draw it in the terminal as ASCII. Fires automatically at turn end in auto-accept modes (acceptEdits, bypassPermissions, auto), or on demand via /checkpoint-diagram. Skips trivial turns.
---

# Checkpoint diagram

At a pause, capture what changed this turn as one flowchart, append it to the day's checkpoint log, and draw it in the terminal so the work can be reviewed without re-reading the transcript.

## When to draw

Draw when the turn produced substantive work: several file edits, a multi-step change, a new module or flow, or a decision taken among alternatives.

Skip (write nothing, say one line saying so) when the turn was a single small edit, a read-only investigation, or a question answered in prose.

## Diagram rules

The checkpoint is drawn in the terminal by `mermaid-ascii`, which renders a subset of Mermaid. Stay inside that subset so the diagram draws both in the terminal and on GitHub:

- Use `flowchart TD` for step-and-decision flows, `flowchart LR` for pipelines or dependency graphs. Do not use other Mermaid diagram types; the terminal renderer only draws flowcharts.
- Nodes are plain boxes only: `A[Concrete action on a target]`. No `{diamond}` shapes, no `((circles))`, no `[[subroutines]]`.
- Edges are ASCII only: `-->` or labeled `-->|yes|`. Never use unicode arrows such as the long arrow glyph.
- Model a decision as a box whose label is the question, with the branches as labeled edges: `D[All rows valid?] -->|yes| E[Insert]` and `D -->|no| F[Collect errors]`.
- No `classDef` and no `:::class`; the renderer prints them as stray boxes. Mark deferred or skipped work with a text prefix: `G[deferred: wire LiveView upload]`.
- Avoid `{}`, `()` and quotes inside labels; keep labels short and concrete. The label sets the box width. Name a concrete action and its target, not "update code".
- Cap at about 12 nodes. Collapse sub-steps and note the count if larger.

## State line (required)

After the diagram, one line stating what is verified versus assumed. Never imply more certainty than the turn earned.

## Formatting

- No em-dashes anywhere in the file. Use plain ASCII punctuation.
- A new file starts with `# Checkpoints <YYYY-MM-DD>`.
- Each section header is `## <n>. <short title>  (<HH:MM>)`, numbered sequentially. Get date and time from `date`.

## Where to write

Append to `<cwd>/.claude/checkpoints/<YYYY-MM-DD>.md`. Ensure `.claude/checkpoints/` is gitignored.

    ## <n>. <short title>  (<HH:MM>)

    ```mermaid
    flowchart TD
        ...
    ```

    <two or three line plain summary>. State: <verified vs assumed>.

## Draw it in the terminal

After appending, render the new diagram and show it inline so it is visible without expanding a tool call:

1. Run `bash ~/.claude/hooks/render-mermaid.sh "<cwd>/.claude/checkpoints/<YYYY-MM-DD>.md"` (a plugin install uses `hooks/render-mermaid.sh` under the plugin root).
2. Include that ASCII output in your reply inside a plain fenced block, so the drawing appears in your message text and not only in the collapsed tool output.

If the render prints a parse complaint, the diagram left the supported subset; simplify it and render again. The file also renders as a full Mermaid diagram in an IDE preview or on GitHub, and `checkpoint-view` (in the repo) opens an interactive, scrollable view of any checkpoint file.

## Worked example

    ## 1. CSV import path  (14:22)

    ```mermaid
    flowchart TD
        A[Add Portfolio.Import.run/1] --> B[Row valid?]
        B -->|yes| C[Cast to Holding struct]
        B -->|no| D[Collect error, skip row]
        C --> E[Bulk insert via Repo.insert_all]
        D --> F[Return counts and errors]
        E --> F
        F --> G[deferred: wire ImportLive upload]
    ```

    Added Portfolio.Import.run/1: validates rows, casts valid ones, bulk-inserts, returns counts plus per-row errors. State: compiles and unit-tested on happy and malformed-row paths; deferred upload path unbuilt.
