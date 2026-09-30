# Prototype e0f7: text renderers

Task `e0f7` builds the `--text` format: the Markdown content drawn with Unicode box characters, the gate and state symbols and tree indentation, sized to a terminal width. This prototype answers its riskiest question: how wide are emoji gate symbols and variation selectors in a box-drawn column, and can a renderer keep the right border straight?

## Run

```
go test ./prototype/e0f7
go run ./prototype/e0f7 -width 100                 # the global tableau
go run ./prototype/e0f7 -width 64 -vs16 strip      # narrower; selectors removed
go run ./prototype/e0f7 -probe                     # alignment probe for your terminal
```

Flags: `-in` (view data), `-width` (default `$COLUMNS`, else 80), `-vs16 wide|narrow|strip`, `-probe`.

## Test data

`testdata/global-tableau.json` is a copy of `/home/nbyoung/Projects/Tableaux/tablo/prototype/886d/output/global-tableau.json`, which tablo prototype 886d writes with `go run ./prototype/886d tableau` over the corpus's `weather-station` entry. The prototype reads `columns`, `rows` (id, title, depth, cells) and `folded`.

## What it shows

```
┌──────┬───────────────────┬────┬────┬──────┬──────┬────┬────┬────┬────┬──────┐
│ Id   │ Task              │ ❔ │ 📝 │  📌  │  ⚙️  │ ⚡ │ ⚓ │ 📐 │ 🛠️ │  🧩  │
├──────┼───────────────────┼────┼────┼──────┼──────┼────┼────┼────┼────┼──────┤
│ a1c0 │ Weather station   │    │ 🟢 │  🧑  │  🧑  │ 🧑 │ 🧑 │ 🧑 │ 🧑 │  🧑  │
│ 4e2b │   Sensor node     │    │ 🧑 │  🧑  │ 🔴⛔ │ 🧑 │ 🧑 │ 🧑 │ 🧑 │  🧑  │
│ 9f31 │     Sensor board  │    │ 🧑 │ 🧑👀 │ 🔴⛔ │ 🧑 │ 🧑 │ 🧑 │ 🧑 │  🧑  │
│ c07d │     Node firmware │    │ 🧑 │  🧑  │  🧑  │ 🧑 │ —  │ 🟢 │ 🪆 │ 🤖👀 │
```

The findings:

- Two kinds of symbol exist. Symbols with default emoji presentation (📌 ⚡ ⚓ 🟢 🧑 ⛔ 👀 🪆 🤖) take two cells in every terminal that follows Unicode East Asian Width. Symbols written with U+FE0F (⚙️ 🛠️, and 🖼️ in the umbrella example) have a narrow base, so a width table alone gives one cell while the emoji form draws in two.
- Terminals disagree on the selector. Modern ones (kitty, WezTerm, iTerm2, Windows Terminal, recent VTE) widen the cluster to two cells. Older wcwidth-based ones keep one and draw the glyph wide, which pushes the right border out. The prototype implements the three policies as a model: `wide` (selector widens), `narrow` (selector has no width), `strip` (the renderer removes the selector, so every terminal shows the one-cell text glyph).
- With the right model every line of the table has the same width at every tested terminal width (tests run all three models at 120, 100 and 80 cells). The width function is about 100 lines of standard library code: zero width for selectors, joiners and combining marks; two for emoji and East Asian Wide ranges; one otherwise; a selector adds a cell to a narrow base under `wide`.
- The model cannot be detected from inside the process without querying the terminal (a cursor position report after a probe write). `-probe` prints each symbol between bars so a person sees in seconds whether their terminal agrees.
- Width sizing: the task column takes all the shrinkage, truncating with an ellipsis down to 8 cells; the gate columns never shrink. At 64 cells the table is 70 wide, so below about 70 cells the renderer needs a fallback (see below).
- Tree depth shows as two spaces per level in the task column.

## What it leaves out

- Only the global tableau. The other views (queue, blockage tree, contextual tableau, structural and temporal views) need the same width functions and a layout each.
- Bold for parents, dates and notes, the legend, provenance, row and cell states beyond the symbols the view data holds.
- Text wrapping of long cells; multi-line cells; real terminal size detection (`$COLUMNS` only; no `ioctl`).
- Word-wide grapheme clusters beyond selectors and joiners: flags, skin tones and ZWJ sequences count each emoji code point separately, which overstates them.
- The width table is hand-picked for the symbols in the corpus. A release needs a generated table from Unicode data, or a vetted dependency; the module has none.

## What the design gate must decide

1. The default selector policy: `strip` is safe everywhere but loses the emoji look of ⚙️ and 🛠️; `wide` is right on current terminals and breaks older ones; or a `--vs16` flag with a default chosen from `$TERM_PROGRAM` or `$WT_SESSION`. A fourth option changes the symbols: the gate definition picks symbols with default emoji presentation only, which no selector affects.
2. Whether the width table is hand-kept, generated from Unicode data, or taken from a library (this adds a dependency to a module that has only tablo).
3. The behaviour below the minimum width: wrap the table into per-task blocks, drop columns outside the next-gate window, or print at full width and let the terminal wrap.
4. Whether `--text` stays plain (no ANSI colour) for logs, with colour as a separate flag.
5. Layout choices VIEWS.md leaves open, for the mockups: Id column first and left-aligned; centred gate cells; two-space tree indent rather than branch glyphs; single-line box characters; the folded count as a footer line; the header showing symbols only, without gate names.
