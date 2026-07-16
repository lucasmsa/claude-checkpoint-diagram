---
name: checkpoint-diagram
description: At a pause, append a Mermaid checkpoint of this turn's work to .claude/checkpoints/<date>.md AND draw it in the terminal as ASCII. Fires automatically at turn end in auto-accept modes (acceptEdits, bypassPermissions, auto), or on demand via /checkpoint-diagram. Skips trivial turns.
---

# Checkpoint diagram

At a pause, capture what changed this turn as one flowchart, append it to the day's checkpoint log, and draw it in the terminal so the work can be reviewed without re-reading the transcript.

## When to draw

Draw when the turn produced substantive work: several file edits, a multi-step change, a new module or flow, or a decision taken among alternatives.

Skip silently when the turn was a single small edit, a read-only investigation, or a question answered in prose. Produce no output about checkpoints: do not draw, and do not announce that you skipped.

## Diagram rules

The checkpoint is drawn in the terminal by `mermaid-ascii`, which renders a subset of Mermaid. Stay inside that subset so the diagram draws both in the terminal and on GitHub:

- Use `flowchart TD` for step-and-decision flows, `flowchart LR` for pipelines or dependency graphs. Do not use other Mermaid diagram types; the terminal renderer only draws flowcharts.
- Nodes are plain boxes only: `A[Concrete action on a target]`. No `{diamond}` shapes, no `((circles))`, no `[[subroutines]]`.
- Edges are ASCII only: `-->` or labeled `-->|yes|`. Never use unicode arrows such as the long arrow glyph.
- Model a decision as a box whose label is the question, with the branches as labeled edges: `D[All rows valid?] -->|yes| E[Insert]` and `D -->|no| F[Collect errors]`.
- No `classDef` and no `:::class`; the renderer prints them as stray boxes. Mark deferred or skipped work with a text prefix: `G[deferred: wire LiveView upload]`.
- Avoid `{}`, `()` and quotes inside labels; keep labels short and concrete. The label sets the box width. Name a concrete action and its target, not "update code".
- Keep it narrow. The terminal scrolls vertically but not horizontally, so width is the real constraint. Hold labels under about 35 characters (abbreviate: `update: validate + lock + save`, not a sentence); each label sets its box width.
- Limit fan-out to at most three branches from any one node. If a step has more parallel parts, group them into one node or chain them. Prefer a tall vertical spine over a wide fan-out: a long diagram is fine, a wide one is not.
- Use one starting node and a single vertical spine. Do not create several independent root nodes; mermaid-ascii places each root in its own column, which is what makes a diagram sprawl sideways. Chain parallel steps under the one spine instead.
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

After appending, you MUST render the diagram with the helper and paste its exact output. Never hand-draw a tree, never summarize the diagram in prose, never just say it was logged: the drawing the user sees must be the verbatim renderer output.

1. Run `bash ~/.claude/hooks/render-mermaid.sh "<cwd>/.claude/checkpoints/<YYYY-MM-DD>.md"` (a plugin install uses `hooks/render-mermaid.sh` under the plugin root).
2. Copy its stdout verbatim into your reply inside a plain fenced block, so the box drawing appears in your message text and not only in the collapsed tool output.

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
