# rice-rail Brand Colors

## Primary

| Name | Hex | RGB | Usage |
|------|-----|-----|-------|
| Near Black | `#111110` | 17, 17, 16 | Primary dark background, terminal bg |
| Pale | `#F3F1ED` | 243, 241, 237 | Primary light text, light backgrounds |
| Signal Amber | `#D97757` | 217, 119, 87 | THE accent — convergence points, active states, CTA |

## Neutral

| Name | Hex | RGB | Usage |
|------|-----|-----|-------|
| Graphite | `#222220` | 34, 34, 32 | Deep background layers |
| Concrete | `#373633` | 55, 54, 51 | Track bed, borders, separators |
| Oxidized | `#605E59` | 96, 94, 89 | Secondary text, muted elements |
| Stone | `#A6A49E` | 166, 164, 158 | Body text on dark, labels |

## Signal States

| Name | Hex | RGB | Usage |
|------|-----|-----|-------|
| Signal Clear | `#789D5F` | 120, 157, 95 | PASS, success, clear state |
| Signal Amber | `#D97757` | 217, 119, 87 | WARNING, active, attention |
| Signal Stop | `#C4504A` | 196, 80, 74 | FAIL, error, blocking |

## Terminal (ANSI 256)

For terminal output where exact hex isn't available:

| Color | ANSI 256 | ANSI Escape |
|-------|----------|-------------|
| Amber accent | 173 | `\033[38;5;173m` |
| Green pass | 107 | `\033[38;5;107m` |
| Red fail | 167 | `\033[38;5;167m` |
| Stone text | 145 | `\033[38;5;145m` |
| Dim text | 240 | `\033[38;5;240m` |
| Reset | - | `\033[0m` |

## Typography

| Context | Font | Fallback |
|---------|------|----------|
| Logo / wordmark | JetBrains Mono Bold | monospace |
| Tagline | Jura Light | sans-serif |
| Terminal output | System monospace | - |
| Documentation | GeistMono | monospace |

## ASCII Banner

```
    ╲╲╲╲╲╲╲╲╲
     ╲╲╲╲╲╲╲╲──────╮
      ╲╲╲╲╲╲╲──────┤
       ╲╲╲╲╲╲──────┤   ╭──────────────╮
        ╲╲╲╲╲──────┼───┤  rice-rail   │
       ╱╱╱╱╱╱──────┤   ╰──────────────╯
      ╱╱╱╱╱╱╱──────┤
     ╱╱╱╱╱╱╱╱──────╯
    ╱╱╱╱╱╱╱╱╱
```

## Design Concept

**Convergent Precision** — parallel tracks converging at a single junction point. The visual metaphor is railway switching diagrams: many inputs (repo state, tools, standards) merge through a convergence point (the constitution) into focused, controlled output (the toolkit cycle). The amber accent marks the convergence point — the critical path.
