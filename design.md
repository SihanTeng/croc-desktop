# Design — croc GUI

A locked design system for the croc desktop app. Every view (Send / Receive /
Relay / Settings) reads this file before emitting code. Do not regenerate per
view — extend or amend this file when the system needs to grow.

Inferences (auto-approved brief): audience = developers moving files between
machines · use case = complete a transfer, fast · tone = technical, calm
instrument panel.

## Genre

modern-minimal

## Macrostructure family

- App pages: **Workbench** — small functional headings, the UI itself is the
  content. No marketing sections, no hero copy, no eyebrow tags.
- All four views share one shape: left side-rail (N3) + single-column content,
  forms and status surfaces composed with hairlines.

## Theme — Cobalt (catalog)

Cool engineered paper, one electric cobalt signal, hairline structure, tight
6 px radii. Cobalt is a light theme; its one dark beat is the graphite
code-hero card (the transfer code display).

- `--color-paper`      oklch(98.5% 0.004 250)
- `--color-paper-2`    oklch(96.5% 0.005 250)
- `--color-ink`        oklch(24% 0.02 258)
- `--color-ink-2`      oklch(34% 0.018 257)
- `--color-muted`      oklch(52% 0.012 257)
- `--color-rule`       oklch(90% 0.006 250)
- `--color-rule-2`     oklch(84% 0.008 252)
- `--color-accent`     oklch(54% 0.20 256)
- `--color-accent-ink` oklch(99% 0.003 256)
- `--color-focus`      oklch(54% 0.20 256)
- `--color-graphite`   oklch(22% 0.016 260)
- `--color-graphite-2` oklch(30% 0.014 260)
- `--color-ok`         oklch(50% 0.17 155)
- `--color-error`      oklch(55% 0.20 25)

## Typography

- Display: Space Grotesk, weight 600, style normal, tracking -0.02em
- Body:    Inter, weight 400/500
- Mono:    JetBrains Mono, weight 400/500/600 — code phrases, field labels
  (UPPERCASE, 0.06em), status readouts, file sizes
- Fonts are self-hosted in `src/assets/fonts/` (desktop app — must work
  offline). Scale anchor: 1.25 major third on a 15px UI body.

## Spacing

4-point named scale (`--space-3xs` … `--space-4xl`). Values live in
`tokens.css`. Views reference named tokens, never raw values.

## Motion

- Easings: `--ease-out` cubic-bezier(0.16, 1, 0.3, 1) · `--ease-in`
  cubic-bezier(0.7, 0, 0.84, 0) · `--ease-in-out` cubic-bezier(0.65, 0, 0.35, 1)
- Durations: `--dur-micro` 120ms · `--dur-short` 220ms · `--dur-long` 420ms
- Three primitives only: CTA hover lift · one orchestrated view-load reveal ·
  functional progress transition
- Reduced-motion: spatial motion collapses to ≤150ms opacity crossfade

## Microinteractions stance

- Silent success — no "Saved!" toasts; the control itself confirms
- Focus rings instant, 2px cobalt, 2px offset, `:focus-visible` only
- Border thickness constant across all input states (no layout shift)
- Copy-to-clipboard feedback = button label swap, no toast

## CTA voice

- Primary: solid cobalt fill, `--color-accent-ink` text, 6px radius, verb
  labels ("Send files", "Receive", "Save settings")
- Secondary: hairline-bordered ghost, 6px radius, ink text
- Danger: solid error-red fill, 6px radius ("Stop relay")

## Logo

Hand-built Tier-B SVG mark: geometric crocodile head (snout slope, knockout
eye, three knockout teeth), drawn on a 64-grid. Solid ink on paper in the
rail; paper on cobalt for the app icon. Source: `src/assets/logo.svg`.

## Per-page allowances

- App pages MUST NOT use marketing enrichment — function carries the view.
- The one dark beat: the graphite code-display card on the Send view.

## What views MUST share

- The logo mark + wordmark in the rail
- Cobalt accent placement: primary CTA, focus rings, active nav indicator,
  progress fill, graphite-card accents (≤ 5% of any view)
- Space Grotesk / Inter / JetBrains Mono trio
- CTA voice (6px radius, solid fills, verb labels)
- Field-label rhythm: mono uppercase label above input

## What views MAY differ on

- Content composition (form vs. status vs. list)
- The Send view's graphite code card (its own dark moment)

## Exports

### tokens.css

Canonical token block — see `frontend/src/tokens.css` (imported by
`styles.css`; every color/font/space/ease references a named token).
