#!/usr/bin/env python3
"""Post-process mermaid-ascii output: round all box corners, and give the one
node whose label is marked with a leading star a double-line border (emphasis).
Reads the render on stdin, writes the styled render to stdout. If no marked
node is found, it just rounds every box."""
import sys

MARK = "★"  # black star, prefixes the key node's label in the mermaid source

raw = sys.stdin.read().rstrip("\n")
lines = raw.split("\n")
width = max((len(l) for l in lines), default=0)
grid = [list(l.ljust(width)) for l in lines]
H = len(grid)

# locate the marker
mark = None
for r in range(H):
    for c in range(width):
        if grid[r][c] == MARK:
            mark = (r, c)
            break
    if mark:
        break

if mark:
    r0, c0 = mark
    # box borders on the label row: nearest vertical bars left and right
    lc = next((c for c in range(c0, -1, -1) if grid[r0][c] in "│├┤"), None)
    rc = next((c for c in range(c0, width) if grid[r0][c] in "│├┤"), None)
    # top and bottom border rows: nearest corner up/down in the left column
    tr = next((r for r in range(r0, -1, -1) if grid[r][lc] in "┌╭"), None) if lc is not None else None
    br = next((r for r in range(r0, H) if grid[r][lc] in "└╰"), None) if lc is not None else None
    if None not in (lc, rc, tr, br):
        hmap = {"─": "═", "┬": "╤", "┴": "╧", "┼": "╬"}
        vmap = {"│": "║", "├": "╟", "┤": "╢", "┼": "╬"}
        grid[tr][lc], grid[tr][rc] = "╔", "╗"
        grid[br][lc], grid[br][rc] = "╚", "╝"
        for c in range(lc + 1, rc):
            for rr in (tr, br):
                grid[rr][c] = hmap.get(grid[rr][c], grid[rr][c])
        for r in range(tr + 1, br):
            for cc in (lc, rc):
                grid[r][cc] = vmap.get(grid[r][cc], grid[r][cc])
        grid[r0][c0] = " "  # drop the marker glyph

# round every remaining single-line corner
round_map = {"┌": "╭", "┐": "╮", "└": "╰", "┘": "╯"}
for r in range(H):
    for c in range(width):
        if grid[r][c] in round_map:
            grid[r][c] = round_map[grid[r][c]]

print("\n".join("".join(row).rstrip() for row in grid))
